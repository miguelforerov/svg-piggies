import * as React from "react"
import { Search, X } from "lucide-react"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupButton,
} from "@/components/ui/input-group"


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

  React.useEffect(() => {
    setValue(defaultValue)
  }, [defaultValue])

  const handleChange = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    setValue(event.target.value)
  }

  const handleClear = () => {
    setValue("")
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
      <InputGroup className="gap-0 flex">
        <InputGroupInput
          {...props}
          type="text"
          value={value}
          onChange={handleChange}
          placeholder={placeholder}
          className={`${className ?? ""}`}
        />
        {value && (
          <InputGroupAddon
            onClick={handleClear}
            align="inline-end"
            className="pr-1"
          >
            <InputGroupButton variant="ghost" size="icon-sm">
              <X />
            </InputGroupButton>
          </InputGroupAddon>
        )}
        <InputGroupAddon  align="inline-end" className="pr-1">
          <InputGroupButton variant="ghost" size="icon-sm">
            <Search />
          </InputGroupButton>
        </InputGroupAddon>
      </InputGroup>
    </form>
  )
}

export default SearchInput
