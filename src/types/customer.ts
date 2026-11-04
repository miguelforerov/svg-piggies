export interface Customer {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  phone?: string;
  stripeCustomerId?: string;
  authUserId?: string;
  marketingConsent: boolean;
  marketingConsentAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CustomerInput {
  email: string;
  firstName?: string;
  lastName?: string;
  phone?: string;
  marketingConsent: boolean;
}
