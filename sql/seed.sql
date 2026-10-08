-- Synthetic deterministic benchmark data; three tables with exactly 1,000,000 rows each.
SET SESSION foreign_key_checks=0;
CREATE TABLE IF NOT EXISTS users (
 id INT UNSIGNED NOT NULL PRIMARY KEY,
 username VARCHAR(40) NOT NULL,
 balance_cents INT UNSIGNED NOT NULL,
 region_id SMALLINT UNSIGNED NOT NULL,
 KEY idx_users_region (region_id)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS profiles (
 user_id INT UNSIGNED NOT NULL PRIMARY KEY,
 bio VARCHAR(100) NOT NULL,
 rating SMALLINT UNSIGNED NOT NULL
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS orders (
 id INT UNSIGNED NOT NULL PRIMARY KEY,
 user_id INT UNSIGNED NOT NULL,
 amount_cents INT UNSIGNED NOT NULL,
 status TINYINT UNSIGNED NOT NULL,
 KEY idx_orders_user (user_id)
) ENGINE=InnoDB;
CREATE TEMPORARY TABLE digits (d INT NOT NULL PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO digits VALUES (0),(1),(2),(3),(4),(5),(6),(7),(8),(9);
CREATE TEMPORARY TABLE nums (n INT NOT NULL PRIMARY KEY) ENGINE=MEMORY;
INSERT INTO nums SELECT a.d*100+b.d*10+c.d FROM digits a CROSS JOIN digits b CROSS JOIN digits c;
INSERT INTO users (id,username,balance_cents,region_id)
 SELECT n, CONCAT('user_',n), MOD(n*37,100000), MOD(n,100)
 FROM (SELECT a.n*1000+b.n+1 AS n FROM nums a CROSS JOIN nums b) q;
INSERT INTO profiles (user_id,bio,rating)
 SELECT n, CONCAT('Synthetic profile for user ',n), MOD(n*13,100)
 FROM (SELECT a.n*1000+b.n+1 AS n FROM nums a CROSS JOIN nums b) q;
-- Multiplication by 97 is a permutation modulo 1,000,000: every user has exactly one order.
INSERT INTO orders (id,user_id,amount_cents,status)
 SELECT n, MOD((n-1)*97,1000000)+1, MOD(n*53,30000)+100, MOD(n,4)
 FROM (SELECT a.n*1000+b.n+1 AS n FROM nums a CROSS JOIN nums b) q;
ANALYZE TABLE users, profiles, orders;
