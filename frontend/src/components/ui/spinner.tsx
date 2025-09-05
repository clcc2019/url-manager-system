import * as React from "react"
import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: "sm" | "md" | "lg"
  tip?: string
}

const Spinner = React.forwardRef<HTMLDivElement, SpinnerProps>(
  ({ className, size = "md", tip, children, ...props }, ref) => {
    const sizeClasses = {
      sm: "h-4 w-4",
      md: "h-6 w-6", 
      lg: "h-8 w-8"
    }

    if (children) {
      return (
        <div ref={ref} className={cn("relative", className)} {...props}>
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
            <div className="flex flex-col items-center space-y-2">
              <Loader2 className={cn("animate-spin", sizeClasses[size])} />
              {tip && <span className="text-sm text-muted-foreground">{tip}</span>}
            </div>
          </div>
          <div className="opacity-50">
            {children}
          </div>
        </div>
      )
    }

    return (
      <div ref={ref} className={cn("flex items-center justify-center", className)} {...props}>
        <div className="flex flex-col items-center space-y-2">
          <Loader2 className={cn("animate-spin", sizeClasses[size])} />
          {tip && <span className="text-sm text-muted-foreground">{tip}</span>}
        </div>
      </div>
    )
  }
)
Spinner.displayName = "Spinner"

export { Spinner }