-- Create databases for each microservice
CREATE DATABASE IF NOT EXISTS user_db;
CREATE DATABASE IF NOT EXISTS product_db;
CREATE DATABASE IF NOT EXISTS order_db;
CREATE DATABASE IF NOT EXISTS payment_db;

-- Create users for each service
CREATE USER IF NOT EXISTS user_service WITH PASSWORD 'user_password';
CREATE USER IF NOT EXISTS product_service WITH PASSWORD 'product_password';
CREATE USER IF NOT EXISTS order_service WITH PASSWORD 'order_password';
CREATE USER IF NOT EXISTS payment_service WITH PASSWORD 'payment_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE user_db TO user_service;
GRANT ALL PRIVILEGES ON DATABASE product_db TO product_service;
GRANT ALL PRIVILEGES ON DATABASE order_db TO order_service;
GRANT ALL PRIVILEGES ON DATABASE payment_db TO payment_service;