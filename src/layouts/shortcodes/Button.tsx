import * as React from "react"

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode
  className?: string
}

const Button = ({
  children,
  className,
  type = "button",
  ...props
}: ButtonProps) => {
  return (
    <button
      type={type}
      className={`inline-block rounded-md bg-primary px-5 py-2 font-semibold text-white transition-colors duration-300 hover:bg-primary/80 hover:no-underline disabled:cursor-not-allowed disabled:opacity-50 ${className ?? ""}`}
      {...props}
    >
      {children}
    </button>
  )
}

export default Button
