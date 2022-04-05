release: go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest && migrate -path migrations -database $DATABASE_URL up
web: bin/item-based-recommendations