import * as React from "react"
import { Search, X } from "lucide-react"
import { Input } from "@/components/store/ui/Input"

interface SearchInputProps
  extends Omit<
    React.ComponentProps<"input">,
    "type" | "value" | "defaultValue" | "onChange"
  > {
  defaultValue?: string
  onSearch?: (value: string) => void
}

const SearchInput = ({
  className,
  placeholder = "Search...",
  defaultValue = "",
  onSearch,
  ...props
}: SearchInputProps) => {
  const [value, setValue] = React.useState(defaultValue)

  const handleChange = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    setValue(event.target.value)
  }

  const handleClear = () => {
    setValue("")
    onSearch?.("")
  }

  const handleSubmit = (
    event: React.SubmitEvent<HTMLFormElement>,
  ) => {
    event.preventDefault()
    onSearch?.(value)
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="relative w-full"
    >
      <Input
        {...props}
        type="text"
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
        
      />

      <div className="absolute inset-y-0 right-2 flex items-center gap-1">
        {value && (
          <button
            type="button"
            onClick={handleClear}
            className="flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Clear search"
          >
            <X className="size-4" />
          </button>
        )}

        <button
          type="submit"
          className="flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label="Search"
        >
          <Search className="size-4" />
        </button>
      </div>
    </form>
  )
}

export default SearchInput
