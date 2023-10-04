disable_mlock = true
    ui = true

    listener "tcp" {
      tls_disable = true
      address = "[::]:8200"
    }
    storage "file" {
      path = "/vault/data"
    }
