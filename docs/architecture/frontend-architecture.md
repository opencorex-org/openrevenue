# Frontend architecture

Each portal uses feature-oriented source folders. Route components orchestrate feature services and schemas; reusable primitives live in `web/packages`; generated API types are the network boundary. TanStack Query owns server state, Zustand is restricted to client state, React Hook Form and Zod own forms, and i18next owns user-facing strings. Business rules remain server-side. Accessibility tests, MSW integration tests, and Playwright journeys protect critical flows.
