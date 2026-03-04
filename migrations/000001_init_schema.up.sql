-- Create Enums
CREATE TYPE user_type AS ENUM ('investor', 'borrower');
CREATE TYPE loan_status AS ENUM ('proposed', 'approved', 'invested', 'disbursed');
CREATE TYPE investment_status AS ENUM ('process', 'paid', 'invested', 'failed');
CREATE TYPE activity_type AS ENUM ('investment', 'disbursement', 'repayment', 'topup', 'withdrawal');
CREATE TYPE ledger_direction AS ENUM ('CREDIT', 'DEBIT');

-- Users Table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    mask_id UUID UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    type user_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Loans Table
CREATE TABLE loans (
    id BIGSERIAL PRIMARY KEY,
    loan_number VARCHAR(50) UNIQUE NOT NULL,
    borrower_id BIGINT NOT NULL REFERENCES users(id),
    description TEXT NOT NULL,
    principal_amount DECIMAL(18, 2) NOT NULL CHECK (principal_amount > 0),
    rate DECIMAL(5, 4) NOT NULL CHECK (rate >= 0),
    roi DECIMAL(5, 4) NOT NULL CHECK (roi >= 0),
    status loan_status NOT NULL DEFAULT 'proposed',
    total_invested DECIMAL(18, 2) NOT NULL DEFAULT 0 CHECK (total_invested >= 0),
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by_employee_id VARCHAR(50),
    visit_proof_url VARCHAR(255),
    disbursed_at TIMESTAMP WITH TIME ZONE,
    disbursed_by_employee_id VARCHAR(50),
    borrower_agreement_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_total_invested_limit CHECK (total_invested <= principal_amount)
);

-- Investments Table
CREATE TABLE investments (
    id BIGSERIAL PRIMARY KEY,
    loan_id BIGINT NOT NULL REFERENCES loans(id),
    investor_id BIGINT NOT NULL REFERENCES users(id),
    amount DECIMAL(18, 2) NOT NULL CHECK (amount > 0),
    status investment_status NOT NULL DEFAULT 'process',
    idempotent_key VARCHAR(100) NOT NULL,
    agreement_letter_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (investor_id, idempotent_key)
);

-- Pockets Table
CREATE TABLE pockets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES users(id),
    balance_investable DECIMAL(18, 2) NOT NULL DEFAULT 0 CHECK (balance_investable >= 0),
    balance_disbursed DECIMAL(18, 2) NOT NULL DEFAULT 0 CHECK (balance_disbursed >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Pocket Ledger Table
CREATE TABLE pocket_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    amount DECIMAL(18, 2) NOT NULL CHECK (amount > 0),
    direction ledger_direction NOT NULL,
    activity_type activity_type NOT NULL,
    reference_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, activity_type, reference_id, direction)
);

-- Index for performance
CREATE INDEX idx_loans_borrower_id ON loans(borrower_id);
CREATE INDEX idx_investments_loan_id ON investments(loan_id);
CREATE INDEX idx_investments_investor_id ON investments(investor_id);
CREATE INDEX idx_pocket_ledger_user_id ON pocket_ledger(user_id);
