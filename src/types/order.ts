export type OrderStatus = "pending" | "fulfilled" | "cancelled";

export type PaymentStatus =
  | "unpaid"
  | "processing"
  | "paid"
  | "failed"
  | "partially_refunded"
  | "refunded";

export interface OrderItem {
  id: string;
  orderId: string;
  productId: string;
  productName: string;
  variantName?: string;
  sku?: string;
  quantity: number;
  unitPriceAmount: number;
  lineTotalAmount: number;
  createdAt: string;
}

export interface Order {
  id: string;
  customerId?: string;
  customerEmail: string;
  orderStatus: OrderStatus;
  paymentStatus: PaymentStatus;
  currency: string;
  subtotalAmount: number;
  discountAmount: number;
  taxAmount: number;
  totalAmount: number;
  amountRefunded: number;
  stripeCheckoutSessionId?: string;
  stripePaymentIntentId?: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
  paidAt?: string;
  fulfilledAt?: string;
  cancelledAt?: string;
}
