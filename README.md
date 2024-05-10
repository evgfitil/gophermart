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

3. **Balances**
    - `user_id`: INT, Primary Key, Foreign Key (References Users.id)
    - `current_balance`: DECIMAL(10, 2), Default 0.00, Not Null
    - `withdrawn_balance`: DECIMAL(10, 2), Default 0.00, Not Null

4. **Withdrawals**
    - `id`: Primary Key, Serial
    - `user_id`: INT, Foreign Key (References Users.id)
    - `order_number`: VARCHAR(255), Unique, Not Null
    - `amount`: DECIMAL(10, 2), Not Null
    - `status`: VARCHAR(255), Not Null
    - `requested_at`: TIMESTAMP WITH TIME ZONE, Default CURRENT_TIMESTAMP

### Relationships

- **Users** to **Orders**: One-to-Many
    - One User can have multiple Orders.
    - Each Order belongs to exactly one User.

- **Users** to **Balances**: One-to-One
    - Each User has exactly one balance record.
    - The balance record tracks the current and withdrawn balances of loyalty points.

- **Users** to **Withdrawals**: One-to-Many
    - One User can initiate multiple withdrawals.
    - Each withdrawal records the details of loyalty points being used to offset the costs of orders.

- **Orders** to **Withdrawals**: Implicit
    - Withdrawals reference Orders via the order_number.
    - This relationship is not enforced by a foreign key constraint but is logically implied by the unique attribute of order_number.

### Sequence diagram 

```mermaid
```

### Points operations

```mermaid
sequenceDiagram
```