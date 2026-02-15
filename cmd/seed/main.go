package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/lguilherme/contas/internal/config"
	"github.com/lguilherme/contas/internal/migrate"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	if err := migrate.Run(cfg.DatabaseURL); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	db, err := config.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Generate bcrypt hash for "senha123"
	hash, err := bcrypt.GenerateFromPassword([]byte("senha123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to generate password hash: %v", err)
	}
	passwordHash := string(hash)

	log.Println("Seeding users...")
	_, err = db.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash) VALUES
			('a0000000-0000-0000-0000-000000000001', 'Alice', 'alice@example.com', $1),
			('a0000000-0000-0000-0000-000000000002', 'Bob', 'bob@example.com', $1),
			('a0000000-0000-0000-0000-000000000003', 'Carol', 'carol@example.com', $1)
		ON CONFLICT (email) DO NOTHING
	`, passwordHash)
	if err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	log.Println("Seeding household...")
	_, err = db.Exec(ctx, `
		INSERT INTO households (id, name, invite_code) VALUES
			('b0000000-0000-0000-0000-000000000001', 'Casa da Rua das Flores', 'FLORES2024')
		ON CONFLICT (invite_code) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("Failed to seed household: %v", err)
	}

	log.Println("Seeding household members...")
	_, err = db.Exec(ctx, `
		INSERT INTO household_members (household_id, user_id, salary_cents, role) VALUES
			('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 500000, 'admin'),
			('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 350000, 'member'),
			('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 280000, 'member')
		ON CONFLICT (household_id, user_id) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("Failed to seed members: %v", err)
	}

	log.Println("Seeding categories...")
	_, err = db.Exec(ctx, `
		INSERT INTO categories (id, household_id, name, icon) VALUES
			('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Aluguel', '🏠'),
			('c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Energia', '⚡'),
			('c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Internet', '🌐'),
			('c0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001', 'Água', '💧'),
			('c0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001', 'Mercado', '🛒'),
			('c0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000001', 'Gás', '🔥')
		ON CONFLICT (household_id, name) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}

	log.Println("Seeding fixed bills...")
	_, err = db.Exec(ctx, `
		INSERT INTO fixed_bills (household_id, category_id, description, amount_cents, due_day, is_shared, paid_by) VALUES
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'Aluguel mensal', 200000, 5, true, 'a0000000-0000-0000-0000-000000000001'),
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000003', 'Internet fibra', 12990, 10, true, 'a0000000-0000-0000-0000-000000000001'),
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000006', 'Gás encanado', 8500, 15, true, 'a0000000-0000-0000-0000-000000000001')
	`)
	if err != nil {
		log.Fatalf("Failed to seed fixed bills: %v", err)
	}

	log.Println("Seeding sample expenses...")
	_, err = db.Exec(ctx, `
		INSERT INTO expenses (household_id, category_id, description, amount_cents, expense_date, is_shared, paid_by) VALUES
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000005', 'Compras do mês', 45090, '2026-02-01', true, 'a0000000-0000-0000-0000-000000000001'),
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000002', 'Conta de luz fevereiro', 18750, '2026-02-05', true, 'a0000000-0000-0000-0000-000000000002'),
			('b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000004', 'Conta de água', 9500, '2026-02-08', true, 'a0000000-0000-0000-0000-000000000003')
	`)
	if err != nil {
		log.Fatalf("Failed to seed expenses: %v", err)
	}

	log.Println("✅ Seed completed successfully!")
	log.Println("Users: alice@example.com, bob@example.com, carol@example.com (password: senha123)")
	log.Println("Household: Casa da Rua das Flores (invite code: FLORES2024)")
	log.Println("Salaries: Alice R$5000, Bob R$3500, Carol R$2800")
}
