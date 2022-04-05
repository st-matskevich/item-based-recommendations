wget -O /tmp/migrate.linux-amd64.tar.gz https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz &&
tar -xvf /tmp/migrate.linux-amd64.tar.gz -C /tmp &&
/tmp/migrate.linux-amd64 -database "${DATABASE_URL}" -path migrations up