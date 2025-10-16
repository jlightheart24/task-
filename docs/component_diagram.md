flowchart LR
  subgraph Clients
    A1[iOS App (Flutter)]
    A2[macOS App (Flutter)]
    A3[Windows App (Flutter)]
  end

  subgraph Internet
    TLS1[(HTTPS/TLS)]
  end

  subgraph Server["API Server (NestJS/TypeScript)"]
    B1[Auth Module\n(Sign in with Apple/Google → JWT/Refresh)]
    B2[Tasks/Projects API\nCRUD + Validation + Rate Limits]
    B3[Sync Endpoint\n/sync?since=<version>]
    B4[Push Gateway\nAPNs/WNS]
    B5[Real-time (opt)\nWebSockets/SSE]
  end

  subgraph DataPlane
    C1[(PostgreSQL)]
    C2[(Redis):::opt]
    C3[(S3/Object Storage):::opt]
  end

  subgraph Providers
    D1[APNs]
    D2[WNS]
    D3[Apple/Google IdP]
  end

  classDef opt fill:#f7f7f7,stroke:#bbb,color:#444;

  A1 -->|HTTPS + JWT| TLS1 --> B1
  A2 -->|HTTPS + JWT| TLS1 --> B1
  A3 -->|HTTPS + JWT| TLS1 --> B1
  B1 --> B2
  B2 --> B3
  B2 --> C1
  B3 --> C1
  B4 --> D1
  B4 --> D2
  B1 -->|Verify| D3
  B5 --> C2
  B4 --> C2
  B2 --> C2
  B2 --> C3
