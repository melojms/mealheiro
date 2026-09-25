// Curated Lucide icons available for categories. Keys are lucide kebab-case names
// stored in categories.icon. Import individually to keep the bundle small.
import {
  Award, Baby, Banknote, Bike, Bitcoin, BookOpen, Briefcase, Building2, Bus, Car, CarTaxiFront, Cat, Circle,
  CircleEllipsis, Coffee, Coins, CreditCard, Dog, Droplet, Dumbbell, Film, Flame, Fuel, Gamepad2, Gift,
  GraduationCap, HandCoins, HeartPulse, House, KeyRound, Landmark, Laptop, Music, Package, PartyPopper,
  PiggyBank, Pill, Plane, Receipt, Repeat, Scissors, Shirt, ShieldCheck, ShoppingBag, ShoppingCart, Smartphone,
  Sparkles, Stethoscope, TrainFront, TrendingUp, Undo2, Utensils, Wallet, Wifi, Wine, Wrench, Zap,
  type LucideIcon,
} from "lucide-react"

export const CATEGORY_ICONS: Record<string, LucideIcon> = {
  award: Award, baby: Baby, banknote: Banknote, bike: Bike, bitcoin: Bitcoin, "book-open": BookOpen,
  briefcase: Briefcase, "building-2": Building2, bus: Bus, car: Car, "car-taxi-front": CarTaxiFront, cat: Cat,
  circle: Circle, "circle-ellipsis": CircleEllipsis, coffee: Coffee, coins: Coins, "credit-card": CreditCard,
  dog: Dog, droplet: Droplet, dumbbell: Dumbbell, film: Film, flame: Flame, fuel: Fuel, "gamepad-2": Gamepad2,
  gift: Gift, "graduation-cap": GraduationCap, "hand-coins": HandCoins, "heart-pulse": HeartPulse, house: House,
  "key-round": KeyRound, landmark: Landmark, laptop: Laptop, music: Music, package: Package,
  "party-popper": PartyPopper, "piggy-bank": PiggyBank, pill: Pill, plane: Plane, receipt: Receipt,
  repeat: Repeat, scissors: Scissors, shirt: Shirt, "shield-check": ShieldCheck, "shopping-bag": ShoppingBag,
  "shopping-cart": ShoppingCart, smartphone: Smartphone, sparkles: Sparkles, stethoscope: Stethoscope,
  "train-front": TrainFront, "trending-up": TrendingUp, "undo-2": Undo2, utensils: Utensils, wallet: Wallet,
  wifi: Wifi, wine: Wine, wrench: Wrench, zap: Zap,
}

export function categoryIcon(name: string): LucideIcon {
  return CATEGORY_ICONS[name] ?? Circle
}
