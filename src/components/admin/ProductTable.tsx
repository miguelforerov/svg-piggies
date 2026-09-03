import { Package } from "lucide-react"

interface Product {
  id: string
  title: string
  price: string
  status: "draft" | "active" | "archived"
}

interface ProductTableProps {
  products: Product[]
}

export function ProductTable({
  products,
}: ProductTableProps) {
  return (
    <div className="rounded-xl border bg-card text-card-foreground shadow-sm">
      <div className="flex items-center justify-between border-b p-6">
        <div>
          <h3 className="text-lg font-semibold">Recent Products</h3>
        </div>
      </div>

      {products.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-2 p-10 text-center">
          <Package className="size-8 text-muted-foreground" />

          <p className="text-sm font-medium">
            No products found
          </p>

          <p className="text-xs text-muted-foreground">
            Products will appear here once they are added.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left">
                <th className="h-12 px-6 font-medium text-muted-foreground">
                  Product
                </th>

                <th className="h-12 px-6 font-medium text-muted-foreground">
                  Price
                </th>

                <th className="h-12 px-6 font-medium text-muted-foreground">
                  Status
                </th>
              </tr>
            </thead>

            <tbody>
              {products.map((product) => (
                <tr
                  key={product.id}
                  className="border-b last:border-0"
                >
                  <td className="px-6 py-4 font-medium">
                    {product.title}
                  </td>

                  <td className="px-6 py-4">
                    {product.price}
                  </td>

                  <td className="px-6 py-4">
                    <span
                      className={[
                        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
                        product.status === "active" &&
                          "bg-primary/10 text-primary",
                        product.status === "draft" &&
                          "bg-muted text-muted-foreground",
                        product.status === "archived" &&
                          "bg-destructive/10 text-destructive",
                      ]
                        .filter(Boolean)
                        .join(" ")}
                    >
                      {product.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}