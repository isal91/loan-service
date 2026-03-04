-- Seed Initial Data (Dummy Borrower for Testing)
INSERT INTO users (id, mask_id, name, username, password, type) 
VALUES (1, '550e8400-e29b-41d4-a716-446655440000', 'Dummy Borrower', 'borrower1', 'password123', 'borrower')
ON CONFLICT (username) DO NOTHING;

-- Seed Investor
INSERT INTO users (id, mask_id, name, username, password, type)
VALUES 
	(2, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Dummy Investor 1', 'investor1', 'password123', 'investor'),
	(3, 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'Dummy Investor 2', 'investor2', 'password123', 'investor'),
	(4, 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'Dummy Investor 3', 'investor3', 'password123', 'investor')
ON CONFLICT (username) DO NOTHING;

-- Seed Pockets
INSERT INTO pockets (user_id, balance_investable)
VALUES 
	(1, 0), 
	(2, 100000000),
	(3, 100000000),
	(4, 100000000)
ON CONFLICT (user_id) DO UPDATE SET balance_investable = EXCLUDED.balance_investable;
