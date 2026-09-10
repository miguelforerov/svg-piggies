import { Avatar } from "./Avatar";

export default function NavUser() {
  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-primary">Account</span>
      <Avatar name="Miguel Forero" className="size-8" />
    </div>
  );
}
