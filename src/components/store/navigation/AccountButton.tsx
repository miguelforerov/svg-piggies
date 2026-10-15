import { useState, type FormEvent } from "react"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp"
import { Label } from "@/components/ui/label"
import { REGEXP_ONLY_DIGITS } from "input-otp"
import Profile from "@/components/store/icons/Profile"

interface AccountButtonProps {
  isAuthenticated?: boolean
}

export default function AccountButton({
  isAuthenticated = false,
}: AccountButtonProps) {
  const [step, setStep] = useState<"email" | "otp">("email")
  const [email, setEmail] = useState("")
  const [otp, setOtp] = useState("")
  const [message, setMessage] = useState("")

  function handleSignIn(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (step === "email") {
      setStep("otp")
      setMessage("")
      return
    }

    setMessage("Code verification is not available yet.")
  }

  function handleChangeEmail() {
    setStep("email")
    setOtp("")
    setMessage("")
  }

  function handleResendCode() {
    setMessage("Code resend is not available yet.")
  }

  if (!isAuthenticated) {
    return (
      <Dialog>
        <DialogTrigger
          aria-label="Sign in"
          className="flex size-6 cursor-pointer text-primary hover:text-primary/70 duration-300"
        >
          <Profile />
        </DialogTrigger>
        <DialogContent showDialogFooter={false} className="flex flex-col w-[25rem] py-[3rem] items-center justify-center">
          <div className="flex flex-col w-[18rem] gap-6">
            <div className="flex flex-col gap-2">
              <DialogTitle className="text-center !text-heading-2">
                {step === "otp" ? "Check your email" : "Sign in"}
              </DialogTitle>
              <DialogDescription className="text-center">
                {step === "otp"
                  ? `We've sent a sign-in code to your email address. Enter the code below to continue.`
                  : "Enter your email address and we'll send you a one-time code to sign in."}
              </DialogDescription>
            </div>
            <form className="flex flex-col gap-4" onSubmit={handleSignIn}>
              <Label htmlFor="sign-in-email" className="sr-only">
                Email address
              </Label>
              <Input
                id="sign-in-email"
                name="email"
                type="email"
                placeholder="Email address"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                disabled={step === "otp"}
                required
              />
              {step === "otp" && (
                <>
                  <Label htmlFor="sign-in-otp" className="sr-only">Access code</Label>
                  <InputOTP
                    id="sign-in-otp"
                    name="otp"
                    maxLength={6}
                    minLength={6}
                    pattern={REGEXP_ONLY_DIGITS}
                    autoComplete="one-time-code"
                    containerClassName="justify-center"
                    value={otp}
                    onChange={(value) => {
                      setOtp(value)
                      setMessage("")
                    }}
                    required
                  >
                    <InputOTPGroup className="w-full justify-center">
                      <InputOTPSlot index={0} />
                      <InputOTPSlot index={1} />
                      <InputOTPSlot index={2} />
                      <InputOTPSlot index={3} />
                      <InputOTPSlot index={4} />
                      <InputOTPSlot index={5} />
                    </InputOTPGroup>
                  </InputOTP>
                </>
              )}
              <Button type="submit" className="w-full">
                {step === "otp" ? "Verify code" : "Send Access Code"}
              </Button>
              {step === "otp" ? (
                <div className="flex flex-col gap-2 text-center text-sm text-muted-foreground">
                  <p>
                    Didn't get a code?{" "}
                    <Button
                      variant="link"
                      className="font-normal"
                      onClick={handleResendCode}
                    >
                      Resend code
                    </Button>
                  </p>
                  <p>
                    Use a{" "}
                    <Button
                      variant="link"
                      className="font-normal"
                      onClick={handleChangeEmail}
                    >
                      Different email address
                    </Button>
                  </p>
                </div>
              ) : (
                <p className="text-center text-sm text-muted-foreground">
                  No password required.
                </p>
              )}
              {message && (
                <p role="status" className="text-center text-sm text-muted-foreground">
                  {message}
                </p>
              )}
            </form>
          </div>
        </DialogContent>
      </Dialog>
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
