-- Seed Initial Loan (Dummy Loan for Testing)
INSERT INTO loans (
    loan_number, 
    borrower_id, 
    description, 
    principal_amount, 
    rate, 
    roi, 
    status
) 
VALUES (
    'LN-DUMMY-001', 
    1, 
    'Pinjaman Modal Usaha Bakso', 
    5000000, 
    0.10, 
    0.12, 
    'proposed'
)
ON CONFLICT (loan_number) DO NOTHING;
