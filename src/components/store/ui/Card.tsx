import type { ReactNode } from "react";

interface CardProps {
  children: ReactNode;
  className?: string
}

const Card = ({ children }: CardProps) => (
  <div className="flex flex-col w-full bg-background-main p-6 rounded-[1rem] gap-4">
    {children}
  </div>
);

export default Card;
