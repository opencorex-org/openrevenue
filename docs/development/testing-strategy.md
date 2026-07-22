# Testing strategy

Domain tests cover invariants such as money and reversals. Application tests exercise commands with fake ports. Repository and migration tests run against disposable PostgreSQL containers. Contract tests validate OpenAPI/events and generated clients. React Testing Library covers components/features with MSW; Playwright covers the vertical slice. Architecture tests block forbidden imports/table access. Authorization tests enumerate allow/deny matrices. Security, accessibility, and performance suites gate releases according to risk.
