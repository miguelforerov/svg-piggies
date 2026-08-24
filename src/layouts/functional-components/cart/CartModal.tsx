import { useStore } from "@nanostores/react";
import React, { useState } from "react";
import { FaShoppingCart } from "react-icons/fa";

import {
  $cart,
  $cartCount,
  $cartTotal,
  removeFromCart,
  updateQuantity,
} from "@/cartStore";

import CloseCart from "./CloseCart";
import OpenCart from "./OpenCart";

const CartModal: React.FC = () => {
  const currentCart = useStore($cart);
  const quantity = useStore($cartCount);
  const total = useStore($cartTotal);

  const [isOpen, setIsOpen] = useState(false);

  const openCart = () => {
    setIsOpen(true);
    document.body.style.overflow = "hidden";
  };

  const closeCart = () => {
    setIsOpen(false);
    document.body.style.overflow = "";
  };

  const formatPrice = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(amount);
  };

  return (
    <>
      {/* Cart icon */}
      <div
        className="cursor-pointer"
        aria-label="Open cart"
        onClick={openCart}
      >
        <OpenCart quantity={quantity} />
      </div>

      {/* Overlay */}
      <div
        id="cartOverlay"
        className={`fixed inset-0 z-40 bg-black opacity-50 transition-opacity ${
          isOpen ? "block" : "hidden"
        }`}
        onClick={closeCart}
      />

      {/* Cart drawer */}
      <div
        id="cartDialog"
        className={`fixed inset-y-0 right-0 z-50 w-full transform transition-transform duration-300 ease-in-out md:w-[390px] ${
          isOpen ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <div className="flex h-full flex-col border-l border-neutral-200 bg-body p-6 text-black drop-shadow-lg dark:border-neutral-700 dark:bg-darkmode-body dark:text-white">
          
          {/* Header */}
          <div className="flex items-center justify-between">
            <p className="text-lg font-semibold">Your Cart</p>

            <button aria-label="Close cart" onClick={closeCart}>
              <CloseCart />
            </button>
          </div>

          <div className="absolute left-0 top-16 h-px w-full bg-dark dark:bg-light" />

          {/* Empty cart */}
          {currentCart.length === 0 ? (
            <div className="my-auto flex flex-col items-center justify-center space-y-6">
              <FaShoppingCart size={76} />

              <p>Oops. Your Bag Is Empty.</p>

              <a
                href="/products"
                className="btn btn-primary w-full"
                onClick={closeCart}
              >
                Don't Miss Out: Add Product
              </a>
            </div>
          ) : (
            <>
              {/* Items */}
              <div className="flex h-full flex-col justify-between overflow-hidden p-1">
                <ul className="flex-grow overflow-auto py-4">
                  {currentCart.map((item) => (
                    <li
                      key={item.variantId}
                      className="flex w-full flex-col border-b border-neutral-300 dark:border-neutral-700"
                    >
                      <div className="flex w-full flex-row justify-between px-1 py-4">

                        {/* Product */}
                        <div className="flex flex-row space-x-4">
                          <div className="relative h-16 w-16 overflow-hidden rounded-md border border-neutral-300 bg-neutral-300">
                            <img
                              className="h-full w-full object-cover"
                              src={
                                item.image ||
                                "/images/product-placeholder.jpg"
                              }
                              alt={item.name}
                              width={64}
                              height={64}
                            />
                          </div>

                          <div className="flex flex-1 flex-col text-base">
                            <span>{item.name}</span>

                            {item.variantName && (
                              <p className="text-sm text-neutral-500">
                                {item.variantName}
                              </p>
                            )}

                            <p className="mt-1 text-sm">
                              {formatPrice(item.price)}
                            </p>
                          </div>
                        </div>

                        {/* Quantity + delete */}
                        <div className="ml-1 flex h-16 flex-col items-end justify-between">
                          <button
                            type="button"
                            className="text-sm text-neutral-500 hover:text-black dark:hover:text-white"
                            onClick={() => removeFromCart(item.variantId)}
                            aria-label={`Remove ${item.name}`}
                          >
                            ×
                          </button>

                          <div className="flex items-center space-x-2">
                            <button
                              type="button"
                              className="flex h-6 w-6 items-center justify-center rounded border"
                              onClick={() =>
                                updateQuantity(
                                  item.variantId,
                                  item.quantity - 1,
                                )
                              }
                              aria-label="Decrease quantity"
                            >
                              −
                            </button>

                            <p>{item.quantity}</p>

                            <button
                              type="button"
                              className="flex h-6 w-6 items-center justify-center rounded border"
                              onClick={() =>
                                updateQuantity(
                                  item.variantId,
                                  item.quantity + 1,
                                )
                              }
                              aria-label="Increase quantity"
                            >
                              +
                            </button>
                          </div>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>

                {/* Summary */}
                <div className="py-4 text-sm text-neutral-500 dark:text-neutral-400">
                  <div className="mb-3 flex items-center justify-between border-b border-neutral-200 pb-1 dark:border-neutral-700">
                    <p>Subtotal</p>

                    <p className="text-right text-base text-black dark:text-white">
                      {formatPrice(total)}
                    </p>
                  </div>

                  <div className="mb-3 flex items-center justify-between border-b border-neutral-200 pb-1 pt-1 dark:border-neutral-700">
                    <p>Shipping</p>
                    <p className="text-right">
                      Calculated at checkout
                    </p>
                  </div>

                  <div className="mb-3 flex items-center justify-between border-b border-neutral-200 pb-1 pt-1 dark:border-neutral-700">
                    <p>Total</p>

                    <p className="text-right text-base text-black dark:text-white">
                      {formatPrice(total)}
                    </p>
                  </div>
                </div>

                {/* Checkout */}
                <button
                  type="button"
                  className="block w-full rounded-md bg-dark p-3 text-center text-sm font-medium text-white opacity-100 hover:opacity-90 dark:bg-light dark:text-text-dark"
                  onClick={() => {
                    console.log("Checkout:", currentCart);
                  }}
                >
                  Proceed to Checkout
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  );
};

export default CartModal;