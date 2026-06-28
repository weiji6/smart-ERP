CREATE DATABASE IF NOT EXISTS smarterp
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'smarterp'@'%' IDENTIFIED BY 'smarterp_password';
CREATE USER IF NOT EXISTS 'smarterp'@'localhost' IDENTIFIED BY 'smarterp_password';

GRANT ALL PRIVILEGES ON smarterp.* TO 'smarterp'@'%';
GRANT ALL PRIVILEGES ON smarterp.* TO 'smarterp'@'localhost';

FLUSH PRIVILEGES;
