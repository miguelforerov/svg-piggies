interface AvatarProps {
  name: string
  className?: string
}

const backgroundColors = [
  "bg-pink-500",
  "bg-blue-500",
  "bg-green-500",
  "bg-purple-500",
  "bg-orange-500",
  "bg-teal-500",
  "bg-indigo-500",
  "bg-rose-500",
]

function getInitials(name: string) {
  const parts = name.trim().split(/\s+/)

  if (parts.length === 0) {
    return ""
  }

  const firstInitial = parts[0]?.charAt(0) ?? ""
  const lastInitial =
    parts.length > 1 ? parts[parts.length - 1]?.charAt(0) ?? "" : ""

  return `${firstInitial}${lastInitial}`.toUpperCase()
}

function getBackgroundColor(name: string) {
  const hash = name
    .split("")
    .reduce((acc, char) => acc + char.charCodeAt(0), 0)

  return backgroundColors[hash % backgroundColors.length]
}

export function Avatar({ name, className = "" }: AvatarProps) {
  const initials = getInitials(name)
  const backgroundColor = getBackgroundColor(name)

  return (
    <div
      className={`flex size-6 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${backgroundColor} ${className}`}
      aria-label={name}
    >
      {initials}
    </div>
  )
}