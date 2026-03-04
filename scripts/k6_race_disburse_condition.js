import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';

// Configuration for 3 concurrent virtual users (employees) trying to disburse the same loan
export const options = {
    scenarios: {
        disburse_race_condition: {
            executor: 'per-vu-iterations',
            vus: 3,
            iterations: 1,
            maxDuration: '10s',
        },
    },
};

export default function () {
    const employeeId = exec.vu.idInTest;

    const url = 'http://localhost:8080/api/v1/admin/loans/LN-DUMMY-001/disburse';

    // Each simulated employee tries to disburse the loan simultaneously
    // Only 1 should succeed, the others should get a 400 Bad Request (loan not invested / already disbursed)
    // or a 409 Conflict if they hit the idempotent key constraint (though here we use different employee IDs).
    const payload = JSON.stringify({
        disbursed_by_employee_id: `EMP-99${employeeId}`,
        borrower_agreement_url: `https://example.com/agreement-emp${employeeId}.pdf`
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Perform the API request
    const res = http.post(url, payload, params);

    if (res.status !== 200 && res.status !== 400 && res.status !== 409) {
        console.log(`Failed req: ${res.status} - ${res.body}`);
    }

    // Assert that the response is returned and not failing due to panics
    // It's normal for 2 requests to fail with 400/409 in this race-condition scenario
    check(res, {
        'is status 200, 400, or 409': (r) => r.status === 200 || r.status === 400 || r.status === 409,
        'no server error': (r) => r.status !== 500,
    });
}
