import { atom, computed } from "nanostores";
import type { Order } from "./types/order";

export type { Order, OrderItem } from "./types/order";

const CART_TOKEN_STORAGE_KEY = "svgpiggies-cart-token";

export interface CartItem {
  id: string;
  productId: string;
  variantId: string;
  name: string;
  variantName?: string;
  price: number;
  quantity: number;
  image?: string;
  slug?: string;
}

interface ApiCartItem {
  id: string;
  productId: string;
  name: string;
  slug: string;
  price: string;
  quantity: number;
}

interface ApiCart {
  id: string;
  status: "active";
  currency: string;
  items: ApiCartItem[];
  createdAt: string;
  updatedAt: string;
  expiresAt: string;
}

interface CreatedCart {
  token: string;
  cart: ApiCart;
}

interface ApiErrorBody {
  message?: string;
}

export interface CheckoutCustomerInput {
  email: string;
  firstName?: string;
  lastName?: string;
  phone?: string;
}

export class CartApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "CartApiError";
  }
}

export const $cart = atom<CartItem[]>([]);
export const $cartDrawerOpen = atom(false);

export const $cartCount = computed($cart, (items) =>
  items.reduce((total, item) => total + item.quantity, 0),
);

export const $cartTotal = computed($cart, (items) =>
  items.reduce((total, item) => total + item.price * item.quantity, 0),
);

export const $cartSummary = computed($cartTotal, (total) => [
  { label: "Subtotal", value: total },
  { label: "Discount", value: "—" },
  { label: "Taxes", value: "Calculated at checkout" },
  { label: "Shipping", value: "Calculated at checkout" },
  { label: "Total", value: total },
]);

let cartCreation: Promise<string> | null = null;
let mutationQueue: Promise<void> = Promise.resolve();

function apiUrl(path: string): string {
  const baseUrl = import.meta.env.PUBLIC_API_BASE_URL?.replace(/\/$/, "");
  return `${baseUrl ?? ""}${path}`;
}

function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(CART_TOKEN_STORAGE_KEY);
}

function storeToken(token: string): void {
  localStorage.setItem(CART_TOKEN_STORAGE_KEY, token);
}

function removeStoredToken(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(CART_TOKEN_STORAGE_KEY);
}

function applyCart(cart: ApiCart): void {
  $cart.set(
    cart.items.map((item) => {
      const price = Number(item.price);
      if (!Number.isFinite(price)) {
        throw new Error(`Cart product ${item.productId} has an invalid price`);
      }

      return {
        id: item.id,
        productId: item.productId,
        // Products currently have no separate variants, so productId is the
        // stable item key expected by the existing cart components.
        variantId: item.productId,
        name: item.name,
        price,
        quantity: item.quantity,
        slug: item.slug,
      };
    }),
  );
}

async function request<T>(
  path: string,
  init: RequestInit = {},
  token?: string,
): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body) headers.set("Content-Type", "application/json");
  if (token) headers.set("X-Cart-Token", token);

  const response = await fetch(apiUrl(path), { ...init, headers });
  if (!response.ok) {
    let body: ApiErrorBody | null = null;
    try {
      body = (await response.json()) as ApiErrorBody;
    } catch {
      // The status code remains sufficient when the response is not JSON.
    }
    throw new CartApiError(
      body?.message ?? `Cart request failed with status ${response.status}`,
      response.status,
    );
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

async function ensureCartToken(): Promise<string> {
  const storedToken = getStoredToken();
  if (storedToken) return storedToken;
  if (cartCreation) return cartCreation;

  cartCreation = request<CreatedCart>("/api/cart", { method: "POST" })
    .then((created) => {
      storeToken(created.token);
      applyCart(created.cart);
      return created.token;
    })
    .finally(() => {
      cartCreation = null;
    });

  return cartCreation;
}

function enqueueMutation(operation: () => Promise<void>): Promise<void> {
  const result = mutationQueue.then(operation, operation);
  mutationQueue = result.catch(() => undefined);
  return result;
}

export async function addToCart(item: CartItem): Promise<void> {
  return enqueueMutation(async () => {
    const token = await ensureCartToken();
    const cart = await request<ApiCart>(
      "/api/cart/items",
      {
        method: "POST",
        body: JSON.stringify({
          productId: item.productId,
          quantity: item.quantity,
        }),
      },
      token,
    );
    applyCart(cart);
  });
}

export async function removeFromCart(variantId: string): Promise<void> {
  return enqueueMutation(async () => {
    const token = getStoredToken();
    if (!token) {
      $cart.set($cart.get().filter((item) => item.variantId !== variantId));
      return;
    }

    const item = $cart.get().find((candidate) => candidate.variantId === variantId);
    const productId = item?.productId ?? variantId;
    const cart = await request<ApiCart>(
      `/api/cart/items/${encodeURIComponent(productId)}`,
      { method: "DELETE" },
      token,
    );
    applyCart(cart);
  });
}

export async function updateQuantity(
  variantId: string,
  quantity: number,
): Promise<void> {
  if (quantity <= 0) return removeFromCart(variantId);

  return enqueueMutation(async () => {
    const token = await ensureCartToken();
    const item = $cart.get().find((candidate) => candidate.variantId === variantId);
    const productId = item?.productId ?? variantId;
    const cart = await request<ApiCart>(
      `/api/cart/items/${encodeURIComponent(productId)}`,
      {
        method: "PUT",
        body: JSON.stringify({ quantity }),
      },
      token,
    );
    applyCart(cart);
  });
}

export async function clearCart(): Promise<void> {
  return enqueueMutation(async () => {
    const token = getStoredToken();
    if (!token) {
      $cart.set([]);
      return;
    }

    await request<void>("/api/cart", { method: "DELETE" }, token);
    $cart.set([]);
  });
}

export async function loadCart(): Promise<void> {
  const token = getStoredToken();
  if (!token) {
    $cart.set([]);
    return;
  }

  try {
    const cart = await request<ApiCart>("/api/cart", {}, token);
    applyCart(cart);
  } catch (error) {
    if (error instanceof CartApiError && error.status === 404) {
      removeStoredToken();
      $cart.set([]);
      return;
    }
    throw error;
  }
}

/**
 * Creates the customer plus a pending order from the current cart. Prices and
 * line items are calculated by the API. Stripe Checkout should be started from
 * the returned order in the payment integration; the browser never supplies
 * amounts or payment status.
 */
export async function createPendingOrder(
  customer: CheckoutCustomerInput,
): Promise<Order> {
  const token = getStoredToken();
  if (!token) {
    throw new CartApiError("Add at least one product before checkout", 400);
  }

  const order = await request<Order>(
    "/api/checkout/orders",
    { method: "POST", body: JSON.stringify(customer) },
    token,
  );
  removeStoredToken();
  $cart.set([]);
  return order;
}
