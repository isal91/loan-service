package main

import (
	"log"
	"net/http"
	"os"

	delivery "loan-service/internal/delivery/http"
	"loan-service/internal/pkg/logger"
	"loan-service/internal/repository"
	"loan-service/internal/usecase"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize structured logger
	logger.Init()

	// 0. Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 1. Setup DB Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Wire Up Clean Architecture Layers
	txManager := repository.NewTransactionManager(db)
	pocketRepo := repository.NewPocketRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	investmentRepo := repository.NewInvestmentRepository(db)
	ledgerRepo := repository.NewLedgerRepository(db)

	loanUsecase := usecase.NewLoanUsecase(loanRepo, investmentRepo, ledgerRepo, pocketRepo, txManager)
	loanHandler := delivery.NewLoanHandler(loanUsecase)
	adminHandler := delivery.NewAdminHandler(loanUsecase)

	// 3. Setup Routing (Standard Library)
	mux := http.NewServeMux()

	// --- PUBLIC / BORROWER ---
	// Endpoint: POST /api/v1/loans (Propose Loan)
	mux.HandleFunc("POST /api/v1/loans", loanHandler.Propose)

	// --- INVESTOR ---
	// Endpoint: POST /api/v1/loans/{loan_number}/invest (Invest Loan)
	mux.HandleFunc("POST /api/v1/loans/{loan_number}/invest", loanHandler.Invest)

	// --- ADMIN / INTERNAL ---
	// Action Endpoints
	mux.HandleFunc("POST /api/v1/admin/loans/{loan_number}/approve", loanHandler.Approve)
	mux.HandleFunc("POST /api/v1/admin/loans/{loan_number}/disburse", loanHandler.Disburse)

	// Debug View Endpoints (For Reviewer Monitoring)
	mux.HandleFunc("GET /api/v1/admin/debug/loans", adminHandler.DebugGetLoans)
	mux.HandleFunc("GET /api/v1/admin/debug/pockets", adminHandler.DebugGetPockets)
	mux.HandleFunc("GET /api/v1/admin/debug/investments", adminHandler.DebugGetInvestments)
	mux.HandleFunc("GET /api/v1/admin/debug/ledger", adminHandler.DebugGetLedger)

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Loan Service starting on port %s...", port)
	handler := delivery.CORSMiddleware(delivery.RequestIDMiddleware(mux))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
