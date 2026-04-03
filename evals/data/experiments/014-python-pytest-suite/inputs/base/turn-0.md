Create a comprehensive pytest test suite for a user authentication system.

Include:
- Fixtures: test database, test client, sample users (admin, regular, inactive), auth headers
- Registration tests: valid signup, duplicate email, weak password, missing fields, email validation
- Login tests: valid login, wrong password, inactive user, rate limiting, token refresh
- Permission tests: admin-only endpoints, resource ownership, role escalation prevention
- Use parametrize for edge cases, proper assertions with descriptive messages

Use section IDs: fixtures, test-registration, test-login, test-permissions

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.