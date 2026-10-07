// import { Link } from "astro:transitions"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import Profile from "@/components/store/icons/Profile"

interface NavUserProps {
  isAuthenticated?: boolean
}

export default function NavUser({
  isAuthenticated = false,
}: NavUserProps) {
  if (!isAuthenticated) {
    return (
      <a
        href="/sign-in"
        aria-label="Sign in"
        className="w-6 h-6 flex text-primary"
      >
        <Profile />
      </a>
    )
  }

  return (
    <div className="flex items-center">
      <Avatar className="size-8">
        <AvatarFallback>MF</AvatarFallback>
      </Avatar>
    </div>
  )
}
