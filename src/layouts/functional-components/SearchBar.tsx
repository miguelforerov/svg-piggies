import { useEffect, useState } from "react"
import SearchInput from "./SearchInput"

const SearchBar = () => {
  const [defaultValue, setDefaultValue] = useState("")

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search)
    const query = searchParams.get("q")

    if (query) {
      setDefaultValue(query)
    }
  }, [])

  const updateURL = (query: string) => {
    const newURL = query
      ? `/products?q=${encodeURIComponent(query)}`
      : "/products"

    window.location.href = newURL
  }

  return (
    <SearchInput
      name="search"
      id="searchInput"
      placeholder="Search for products"
      autoComplete="off"
      defaultValue={defaultValue}
      onSearch={updateURL}
    />
  )
}

export default SearchBar
