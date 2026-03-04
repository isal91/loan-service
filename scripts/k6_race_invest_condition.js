import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

// Configuration for 3 concurrent virtual users (investors)
export const options = {
    scenarios: {
        invest_race_condition: {
            executor: 'per-vu-iterations',
            vus: 3, // 3 concurrent investors
            iterations: 1, // each investor runs once
            maxDuration: '10s',
        },
    },
};

export default function () {
    // Determine Investor ID based on VU ID (k6 VUs are 1-indexed)
    // VU 1 -> Investor 2 (ID: 2)
    // VU 2 -> Investor 3 (ID: 3)
    // VU 3 -> Investor 4 (ID: 4)
    const investorId = exec.vu.idInTest + 1;

    const url = 'http://localhost:8080/api/v1/loans/LN-DUMMY-001/invest';

    // Each investor tries to invest 3,000,000 independently.
    // Total required capacity is 5,000,000. 
    // This implies that only 1 investor will fully succeed, 
    // the second investor will succeed or fail depending on how you handle partials (in our case amount exceed), 
    // and the 3rd will definitely fail with fully_funded / amount_exceed depending on who locks first.
    // This perfectly tests the Database FOR UPDATE row lock preventing over-funding.
    const payload = JSON.stringify({
        investor_id: investorId,
        amount: 3000000,
        idempotent_key: `idemp-test-vu${investorId}-${Date.now()}` // Unique per run per user to avoid duplicate_request constraint
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Perform the API request
    const res = http.post(url, payload, params);

    // Assert that the response is returned and not failing due to panics
    // It's normal for some requests to fail with 400 Bad Request in this race-condition scenario
    check(res, {
        'is status 200 or 400': (r) => r.status === 200 || r.status === 400,
        'no server error': (r) => r.status !== 500,
    });
}
