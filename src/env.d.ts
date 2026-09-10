/// <reference types="@cloudflare/workers-types" />

declare global {
  namespace Cloudflare {
    interface Env {
      API: Fetcher;
    }
  }
}

export {};
