package web

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rocketb/asperitas/pkg/logger"
)

// Handler is a type that handles an HTTP request.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint of out application.
type App struct {
	mux      *httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
	tracer   trace.Tracer
	log      *logger.Logger
}

// NewApp creates new instance of our application.
func NewApp(shutdown chan os.Signal, tracer trace.Tracer, log *logger.Logger, mw ...Middleware) *App {
	m := httptreemux.NewContextMux()
	m.RedirectTrailingSlash = false

	return &App{
		mux:      m,
		shutdown: shutdown,
		mw:       mw,
		tracer:   tracer,
		log:      log,
	}
}

// SignalShutdown is used to gracefully shut down the app.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// Handle sets a handler function for a given HTTP method.
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.startSpan(w, r)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now().UTC(),
		}

		ctx = SetValues(ctx, &v)

		if err := handler(ctx, w, r); err != nil {
			if validateShutdown(err) {
				a.SignalShutdown()
				return
			}
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.mux.Handle(method, finalPath, h)
}

// EnableCORS enables CORS preflight requests to work in the middleware. It
// prevents the MethodNotAllowedHandler from being called. This must be enabled
// for the CORS middleware to work.
func (a *App) EnableCORS(mw Middleware) {
	a.mw = append(a.mw, mw)

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return Respond(ctx, w, r, http.StatusOK)
	}
	handler = wrapMiddleware(a.mw, handler)

	a.mux.OptionsHandler = func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		ctx, span := a.startSpan(w, r)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now().UTC(),
		}

		ctx = SetValues(ctx, &v)

		if err := handler(ctx, w, r); err != nil {
			a.log.Error(ctx, "handle: %v", err)
		}
	}
}

func (a *App) startSpan(w http.ResponseWriter, r *http.Request) (context.Context, trace.Span) {
	ctx := r.Context()

	span := trace.SpanFromContext(ctx)

	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "pkg.web.Handle")
		span.SetAttributes(attribute.String("endpoint", r.RequestURI))
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		traceID := spanCtx.TraceID()
		hexTraceID := hex.EncodeToString(traceID[:])

		spanID := spanCtx.SpanID()
		hexSpanID := hex.EncodeToString(spanID[:])

		traceParrent := fmt.Sprintf("traceparent,desc=00-%s-%s-01", hexTraceID, hexSpanID)

		w.Header().Add("Access-Control-Expose-Headers", "Server-Timing")
		w.Header().Add("Server-Timing", traceParrent)
	}

	return ctx, span
}

// validateShutdown validates the error for special conditions that do not
// warrant an actual shutdown by the system.
func validateShutdown(err error) bool {
	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data is being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe error.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP error that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
