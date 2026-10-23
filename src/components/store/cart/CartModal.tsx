import { useStore } from "@nanostores/react";
import React, { useEffect } from "react";
import { FaShoppingCart } from "react-icons/fa";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import {
  $cart,
  $cartCount,
  $cartDrawerOpen,
  $cartSummary,
  loadCart,
  updateQuantity,
} from "@/cartStore";

import CloseIcon from "@/components/store/icons/Close";
import PlusIcon from "@/components/store/icons/Plus";
import TrashIcon from "@/components/store/icons/Trash";
import MinusIcon from "@/components/store/icons/Minus";
import CartIcon from "@/components/store/icons/Cart";
import CardIcon from "@/components/store/icons/Card";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerFooter,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";

const CartModal: React.FC = () => {
  const currentCart = useStore($cart);
  const quantity = useStore($cartCount);
  const summaryItems = useStore($cartSummary);
  const isOpen = useStore($cartDrawerOpen);

  useEffect(() => {
    loadCart();
  }, []);

  const formatPrice = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(amount);
  };

  return (
    <Drawer
      swipeDirection="right"
      open={isOpen}
      onOpenChange={(open) => $cartDrawerOpen.set(open)}
      triggerId="cart-drawer-trigger"
    >
      <DrawerTrigger
        id="cart-drawer-trigger"
        render={
          <Button type="button" variant="link" className="relative text-primary" aria-label="Open cart" />
        }
      >
        <CartIcon filled className="size-8" />
        {quantity > 0 && (
          <span className="absolute -right-3 -top-2 flex size-5 items-center justify-center rounded-full bg-black text-xs text-white">
            {quantity}
          </span>
        )}
      </DrawerTrigger>

      <DrawerContent className="w-[36rem]">
        <div className="flex min-h-0 flex-1 flex-col gap-8">
          {/* Header */}
          <DrawerHeader className="flex-row items-center justify-between">
            <DrawerTitle className="">Shopping Cart</DrawerTitle>

            <DrawerClose aria-label="Close cart" className="cursor-pointer">
              <Button
                variant="link"
                size="icon"
                className="text-muted-foreground"
                aria-label="Close modal"
              >
                <CloseIcon className="size-5" />
              </Button>
            </DrawerClose>
          </DrawerHeader>

          {/* Empty cart */}
          {currentCart.length === 0 ? (
            <div className="my-auto flex flex-col items-center justify-center space-y-6">
              <CartIcon filled className="size-32 text-primary" />

              <p>Oops. Your Bag Is Empty.</p>
              <span className="flex gap-2">
                <p>Don't Miss Out:</p>
                <DrawerClose
                  render={<a href="/products" className="text-primary" />}
                >
                  Add Product
                </DrawerClose>
              </span>
            </div>
          ) : (
            <>
              {/* Items */}
              <div className="flex-1 overflow-y-auto p-4">
                <ul className="flex flex-col min-h-0 flex-1 overflow-y-auto gap-4">
                  {currentCart.map((item) => (
                    <Card
                      key={item.variantId}
                      className="bg-secondary "
                    >
                      <CardContent >
                        <div className="flex w-full flex-row gap-2">
                          <div className="h-[8rem] w-[8rem] overflow-hidden rounded-md flex-none">
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
                          <div className="flex flex-col gap-2 flex-auto">
                            <div className="flex flex-1 flex-col text-lg">
                              <span>{item.name}</span>

                              {item.variantName && (
                                <p className="text-sm text-neutral-500">
                                  {item.variantName}
                                </p>
                              )}
                            </div>
                            <div className="flex w-full justify-between items-ceter">
                              <div className="flex items-center rounded-xl px-2 py-1 bg-background-secondary">
                                <Button
                                  variant="link"
                                  size="icon"
                                  onClick={() =>
                                    updateQuantity(
                                      item.variantId,
                                      item.quantity - 1,
                                    )
                                  }
                                  aria-label={
                                    item.quantity === 1
                                      ? `Remove ${item.name}`
                                      : "Decrease quantity"
                                  }
                                >
                                  {item.quantity === 1 ? (
                                    <TrashIcon className="size-5" />
                                  ) : (
                                    <MinusIcon className="size-5" />
                                  )}
                                </Button>

                                <p>{item.quantity}</p>

                                <Button
                                  variant="link"
                                  size="icon"
                                  onClick={() =>
                                    updateQuantity(
                                      item.variantId,
                                      item.quantity + 1,
                                    )
                                  }
                                  aria-label="Increase quantity"
                                >
                                  <PlusIcon className="size-5" />
                                </Button>
                              </div>
                              <div className="flex items-center gap-2">
                                <p className="text-sm text-muted-foreground line-through">
                                  {formatPrice(item.price)}
                                </p>
                                <p className="text-lg font-bold">
                                  {formatPrice(item.price)}
                                </p>
                              </div>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </ul>
              </div>
              {/* Summary */}
              <DrawerFooter className="gap-8">
                <div className="flex flex-col gap-2">
                  {summaryItems.map(({ label, value }) => (
                    <div
                      key={label}
                      className={`flex items-center justify-between text-base ${label === "Total" && "font-bold border-t-[1px] pt-1"}`}
                    >
                      <p>{label}</p>
                      <p className="text-right">
                        {typeof value === "number" ? formatPrice(value) : value}
                      </p>
                    </div>
                  ))}
                </div>
                {/* Checkout */}
                <div className="space-y-4">
                  <a
                    className={cn(
                      buttonVariants({ variant: "default", size: "2xl" }),
                      "h-auto w-full p-3",
                    )}
                    onClick={() => {}}
                  >
                    <CardIcon className="size-6"/>
                    Proceed to Checkout
                  </a>
                  <DrawerClose aria-label="Close cart" className="w-full">
                    <Button
                      variant="secondary"
                      size="2xl"
                      className="text-muted-foreground py-6 w-full"
                      aria-label="Close modal"
                    >
                      Continue Shopping
                    </Button>
                  </DrawerClose>
                </div>
              </DrawerFooter>
            </>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  );
};

export default CartModal;
