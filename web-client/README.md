# Web Client

This directory contains the Next.js frontend application for the Delayed Wallet Transactions system. It provides a user interface for creating wallets, scheduling transactions, and viewing transaction history.

## Getting Started

### Prerequisites

- Node.js (v20+)
- Yarn

### Setup

1.  **Install dependencies:**

    Navigate to the `web-client` directory and run:
    ```sh
    yarn install
    ```

2.  **Configure Environment Variables:**

    Create a `.env.local` file in the `web-client` directory by copying the example:
    ```sh
    cp .env.local.example .env.local
    ```

    Update `.env.local` with the URL of your backend API service:
    ```
    NEXT_PUBLIC_API_URL=http://localhost:8080
    ```

3.  **Run the Development Server:**

    ```sh
    yarn dev
    ```

    Open [http://localhost:3000](http://localhost:3000) in your browser to see the application.

## Build

To create a production-ready build of the application, run:

```sh
yarn build
```

This will generate a static site in the `out` directory, which can be deployed to any static hosting service.
