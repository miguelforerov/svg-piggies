import * as React from "react"
import { Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import Close from "@/components/store/icons/Close"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
} from "@/components/ui/input-group"
import { ButtonGroup } from "@/components/ui/button-group"


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
            <Button variant="ghost" className=""><Close /></Button>
          </InputGroupAddon>
        )}
        <InputGroupAddon  align="inline-end" className="pr-1">
          <Button variant="ghost" className=""><Search /></Button>
          {/* <Button variant="outline" className="m-1">Search</Button> */}
        </InputGroupAddon>
      </InputGroup>
    </form>
  )
}

export default SearchInput
