#!/bin/bash

# Database migration script for GoConfig Guardian

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
COMMAND=${1:-"up"}
STEPS=${2:-""}

# Load .env file if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Database connection string
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-cfguardian}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"
MIGRATIONS_PATH="db/migrations"

echo -e "${GREEN}🗄️  Database Migration Tool${NC}"
echo "Database: ${DB_NAME}"
echo "Host: ${DB_HOST}:${DB_PORT}"
echo ""

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}❌ migrate command not found${NC}"
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Execute migration command
case "$COMMAND" in
    up)
        echo -e "${YELLOW}⬆️  Running migrations up...${NC}"
        if [ -n "$STEPS" ]; then
            migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" up ${STEPS}
        else
            migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" up
        fi
        echo -e "${GREEN}✅ Migrations applied successfully!${NC}"
        ;;
    
    down)
        echo -e "${YELLOW}⬇️  Running migrations down...${NC}"
        if [ -n "$STEPS" ]; then
            migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" down ${STEPS}
        else
            echo -e "${RED}⚠️  Warning: This will rollback ALL migrations!${NC}"
            read -p "Are you sure? (yes/no): " -r
            if [[ $REPLY == "yes" ]]; then
                migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" down
                echo -e "${GREEN}✅ Migrations rolled back!${NC}"
            else
                echo -e "${YELLOW}Migration cancelled.${NC}"
                exit 0
            fi
        fi
        ;;
    
    force)
        if [ -z "$STEPS" ]; then
            echo -e "${RED}❌ Version number required for force command${NC}"
            echo "Usage: $0 force <version>"
            exit 1
        fi
        echo -e "${YELLOW}🔧 Forcing migration version to ${STEPS}...${NC}"
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" force ${STEPS}
        echo -e "${GREEN}✅ Migration version forced to ${STEPS}${NC}"
        ;;
    
    version)
        echo -e "${YELLOW}📋 Current migration version:${NC}"
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" version
        ;;
    
    create)
        if [ -z "$STEPS" ]; then
            echo -e "${RED}❌ Migration name required${NC}"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        echo -e "${YELLOW}📝 Creating new migration: ${STEPS}${NC}"
        migrate create -ext sql -dir ${MIGRATIONS_PATH} -seq ${STEPS}
        echo -e "${GREEN}✅ Migration files created!${NC}"
        ;;
    
    drop)
        echo -e "${RED}⚠️  Warning: This will DROP the entire database!${NC}"
        read -p "Are you sure? Type 'DROP DATABASE' to confirm: " -r
        if [[ $REPLY == "DROP DATABASE" ]]; then
            migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" drop
            echo -e "${GREEN}✅ Database dropped!${NC}"
        else
            echo -e "${YELLOW}Operation cancelled.${NC}"
            exit 0
        fi
        ;;
    
    *)
        echo -e "${RED}❌ Unknown command: $COMMAND${NC}"
        echo ""
        echo "Usage: $0 <command> [steps]"
        echo ""
        echo "Commands:"
        echo "  up [N]       Apply all or N migrations"
        echo "  down [N]     Rollback all or N migrations"
        echo "  force <V>    Force migration version to V"
        echo "  version      Show current migration version"
        echo "  create <name> Create new migration files"
        echo "  drop         Drop entire database"
        echo ""
        echo "Examples:"
        echo "  $0 up              # Apply all pending migrations"
        echo "  $0 up 1            # Apply next 1 migration"
        echo "  $0 down 1          # Rollback 1 migration"
        echo "  $0 create add_users_table  # Create new migration"
        echo "  $0 version         # Show current version"
        exit 1
        ;;
esac

