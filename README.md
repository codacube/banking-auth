# Auth Module for Banking API

## Summary of the auth packages

- app: this is the startup/bootstrap layer. It sets up the HTTP router, loads environment config, creates the DB client, and registers the auth endpoints. See `app.go`.
  - It wires the service into routes like login, verify, and a placeholder register endpoint.

- domain: this is the core auth/business layer. It contains the repository contract, login model, JWT claims/token logic, and role/permission rules that determine which routes a user can access. The main implementation is in `authRepository.go`, with supporting auth-related types under the same package.

- dto: this is the data-transfer layer. It defines the JSON request/response shapes used between clients and the service, such as login payloads and token-related payloads. The request model is defined in `loginRequest.go`.

- service: this is the use-case/business logic layer. It implements auth operations like:
  - login
  - token verification
  - refresh-token flow
  - permission checks against the role system
  - This logic lives in `authService.go`.

In short, the package flow is:

app -> service -> domain
and dto is the contract layer between HTTP and service code.

## User Roles/Routes

**Admin can do the following:**
| GetAllCustomers | GET | /customer |
| GetCustomer | GET | /customer/{customer_id} |
| NewAccount | POST | /customer/{customer_id}/account |
| NewTransaction | POST | /customer/{customer_id}/account/{account_id} |

**User role can do the following:**
| Get Customer by ID | GET | /customer/{customer_id} |
| New Transaction | POST | /customer/{customer_id}/account/{account_id} |

```json
{
  "admin": ["GetAllCustomers", "GetCustomer", "NewAccount". "NewTransaction"],
  "user": ["GetCustomer", "NewTransaction"]
}
```

### Verification of the token

1. Validity of the token (includes the expiry and the signature verification)
2. Verify if the role has access to the resource
3. Need to verify if the resource being accessed is for the same user

## Status Codes Returned

401 - authentication error (credentials do not match)
403 - authorisation error (access is forbidden, no permission for resource)
500 - internal server error ()
