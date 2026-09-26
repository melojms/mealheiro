import type { ReactNode } from "react"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Drawer, DrawerContent, DrawerDescription, DrawerFooter, DrawerHeader, DrawerTitle } from "@/components/ui/drawer"
import { cn } from "@/lib/utils"
import { useIsDesktop } from "./hooks"

interface ResponsiveDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: ReactNode
  description?: ReactNode
  children: ReactNode
  footer?: ReactNode
  className?: string
}

/** Bottom drawer on phones, centered dialog on desktop. The body scrolls; the footer stays visible. */
export function ResponsiveDialog({ open, onOpenChange, title, description, children, footer, className }: ResponsiveDialogProps) {
  const desktop = useIsDesktop()

  if (desktop) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className={cn("flex max-h-[90svh] flex-col sm:max-w-lg", className)}>
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription className={cn(!description && "sr-only")}>{description ?? title}</DialogDescription>
          </DialogHeader>
          <div className="-mx-4 min-h-0 flex-1 overflow-y-auto px-4 py-1">{children}</div>
          {footer && <DialogFooter className="sm:justify-between">{footer}</DialogFooter>}
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} repositionInputs={false}>
      <DrawerContent className="data-[vaul-drawer-direction=bottom]:max-h-[92svh]">
        <DrawerHeader className="text-left">
          <DrawerTitle className="text-lg">{title}</DrawerTitle>
          <DrawerDescription className={cn(!description && "sr-only")}>{description ?? title}</DrawerDescription>
        </DrawerHeader>
        <div className={cn("min-h-0 flex-1 overflow-y-auto px-4 pb-2", className)}>{children}</div>
        {footer && <DrawerFooter className="border-t pb-[max(1rem,env(safe-area-inset-bottom))]">{footer}</DrawerFooter>}
      </DrawerContent>
    </Drawer>
  )
}
