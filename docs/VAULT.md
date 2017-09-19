# Configuration

In order to use breakglass, you need to have your infrastructure configured in a certain way.

Each component that breakglass can get credentials for has some assumptions built in.

# MySQL

In order to get MySQL Credentials with vault, your database must be mounted under `mysql/<hostname>` in vault. Here's an example of how you might do this:

```
vault mount -path=mysql/mysql.example.com database
```

Once the MySQL instance is mounted under the hostname, you'll need to configure it as normal:

```
vault write mysql/mysql.example.com/config/mysql plugin_name=mysql-database-plugin connection_url="vault:vault@tcp(mysql.example.com:3306)/" allowed_roles="readonly"
vault write mysql/mysql.example.com/roles/readonly db_name=mysql creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" default_ttl="1h" max_ttl="24h"
```

