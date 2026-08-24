import { atom, computed } from "nanostores";

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

export const $cart = atom<CartItem[]>([]);

export const $cartCount = computed($cart, (items) =>
  items.reduce((total, item) => total + item.quantity, 0),
);

export const $cartTotal = computed($cart, (items) =>
  items.reduce((total, item) => total + item.price * item.quantity, 0),
);

export function addToCart(item: CartItem) {
  const current = $cart.get();

  const existing = current.find(
    (cartItem) => cartItem.variantId === item.variantId,
  );

  if (existing) {
    $cart.set(
      current.map((cartItem) =>
        cartItem.variantId === item.variantId
          ? {
              ...cartItem,
              quantity: cartItem.quantity + item.quantity,
            }
          : cartItem,
      ),
    );
  } else {
    $cart.set([...current, item]);
  }

  saveCart();
}

export function removeFromCart(variantId: string) {
  $cart.set(
    $cart.get().filter((item) => item.variantId !== variantId),
  );

  saveCart();
}

export function updateQuantity(
  variantId: string,
  quantity: number,
) {
  if (quantity <= 0) {
    removeFromCart(variantId);
    return;
  }

  $cart.set(
    $cart.get().map((item) =>
      item.variantId === variantId
        ? { ...item, quantity }
        : item,
    ),
  );

  saveCart();
}

export function clearCart() {
  $cart.set([]);
  saveCart();
}

function saveCart() {
  if (typeof window === "undefined") return;

  localStorage.setItem(
    "svgpiggies-cart",
    JSON.stringify($cart.get()),
  );
}

export function loadCart() {
  if (typeof window === "undefined") return;

  const saved = localStorage.getItem("svgpiggies-cart");

  if (!saved) return;

  try {
    $cart.set(JSON.parse(saved));
  } catch {
    localStorage.removeItem("svgpiggies-cart");
  }
}