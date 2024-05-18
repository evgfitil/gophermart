# gophermart

## Entity-Relationship Diagram (ERD)

![Gophermart Embeded Diagram](docs/gophermart_erd.drawio.svg)

### Entity-Relationship Diagram (ERD) Description

### Entities and Attributes

1. **Users**
    - `id`: Primary Key, Serial
    - `username`: VARCHAR(255), Unique, Not Null
    - `password_hash`: CHAR(60), Not Null

2. **Orders**
    - `id`: Primary Key, Serial
    - `user_id`: INT, Foreign Key (References Users.id)
    - `order_number`: VARCHAR(255), Unique, Not Null
    - `status`: ENUM('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')
    - `accrual`: DECIMAL(10, 2), Default 0.00
    - `uploaded_at`: TIMESTAMP WITH TIME ZONE, Default CURRENT_TIMESTAMP

3. **Transactions**
    - `id`: Primary Key, Serial
    - `user_id`: INT, Foreign Key (References Users.id)
    - `type`: VARCHAR(10), Not Null -- 'accrual' or 'withdrawal'
    - `amount`: DECIMAL(10, 2), Not Null
    - `order_id`: INT
    - `created_at`: TIMESTAMP WITH TIME ZONE, Default CURRENT_TIMESTAMP

### Relationships

- **Users** to **Orders**: One-to-Many
    - One User can have multiple Orders.
    - Each Order belongs to exactly one User.

- **Users** to **Transactions**: One-to-Many
    - One User can have multiple Transactions.
    - Each Transaction belongs to exactly one User.

### System interaction flow diagram

```mermaid
graph LR
    U[User] -- HTTP Requests --> API[Gophermart API]
    API -- Data Handling --> H[Handlers]
    H -- Process Requests --> DB[(Storage)]
    LP[Loyalty Processor] -- Check Status --> AS[Accrual System]
    AS -- Return Status/Points --> LP
    LP -- Update Order Status --> DB
```

### Sequence diagram

```mermaid
sequenceDiagram
    participant U as User
    participant API as Gophermart API
    participant H as Handlers
    participant DB as Storage
    participant LP as LoyaltyProcessor
    participant AS as Accrual System

    U->>API: POST /api/user/login {username, password}
    API->>H: Authenticate User
    H->>DB: Verify User Credentials
    DB-->>H: Credentials Verified
    H-->>API: Authentication Complete
    API-->>U: Authentication Response

    U->>API: POST /api/user/orders {order_number}
    API->>H: Handle New Order
    H->>DB: Process New Order
    DB-->>H: Order Processed
    H-->>API: Order Response
    API-->>U: Order Status

    Note over LP: Background process
    LP->>DB: Fetch Pending Orders
    DB-->>LP: Pending Orders Listed
    LP->>AS: Query Accrual Status
    AS-->>LP: Accrual Details
    LP->>DB: Update Orders with Accrual Info
    DB-->>LP: Orders Updated

    U->>API: GET /api/user/balance
    API->>H: Query Balance
    H->>DB: Retrieve Balance Information
    DB-->>H: Balance Info
    H-->>API: Balance Response
    API-->>U: Display Balance

    U->>API: POST /api/user/balance/withdraw {order, sum}
    API->>H: Handle Withdrawal
    H->>DB: Process Withdrawal
    DB-->>H: Withdrawal Processed
    H-->>API: Withdrawal Response
    API-->>U: Withdrawal Status

    U->>API: GET /api/user/orders
    API->>H: Fetch Order Details
    H->>DB: Retrieve Order Details
    DB-->>H: Order Details Retrieved
    H-->>API: Order Details Response
    API-->>U: Display Orders
```

