import type { Customer, CustomerInput } from "../types/customer";
import type { Order, OrderStatus } from "../types/order";

interface ApiErrorBody {
  message?: string;
}

export class AdminCommerceApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "AdminCommerceApiError";
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body) headers.set("Content-Type", "application/json");

  const response = await fetch(path, { ...init, headers });
  if (!response.ok) {
    let body: ApiErrorBody | undefined;
    try {
      body = (await response.json()) as ApiErrorBody;
    } catch {
      // Use the HTTP status when an upstream error is not JSON.
    }
    throw new AdminCommerceApiError(
      body?.message ?? `Request failed with status ${response.status}`,
      response.status,
    );
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export function getCustomers(): Promise<Customer[]> {
  return request<Customer[]>("/api/admin/customers");
}

export function getCustomer(customerId: string): Promise<Customer> {
  return request<Customer>(`/api/admin/customers/${encodeURIComponent(customerId)}`);
}

export function createCustomer(input: CustomerInput): Promise<Customer> {
  return request<Customer>("/api/admin/customers", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateCustomer(
  customerId: string,
  input: CustomerInput,
): Promise<Customer> {
  return request<Customer>(`/api/admin/customers/${encodeURIComponent(customerId)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteCustomer(customerId: string): Promise<void> {
  return request<void>(`/api/admin/customers/${encodeURIComponent(customerId)}`, {
    method: "DELETE",
  });
}

export function getOrders(): Promise<Order[]> {
  return request<Order[]>("/api/admin/orders");
}

export function getOrder(orderId: string): Promise<Order> {
  return request<Order>(`/api/admin/orders/${encodeURIComponent(orderId)}`);
}

export function updateOrderStatus(orderId: string, orderStatus: OrderStatus): Promise<Order> {
  return request<Order>(`/api/admin/orders/${encodeURIComponent(orderId)}`, {
    method: "PUT",
    body: JSON.stringify({ orderStatus }),
  });
}
