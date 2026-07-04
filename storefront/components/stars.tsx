import { Star } from "lucide-react";

// Renders a 0-5 star rating with half-star precision.
export function Stars({ value, size = 16 }: { value: number; size?: number }) {
  const full = Math.floor(value);
  const half = value - full >= 0.5;
  return (
    <span className="inline-flex items-center gap-0.5 align-middle">
      {Array.from({ length: 5 }).map((_, i) => {
        const filled = i < full;
        const isHalf = i === full && half;
        return (
          <Star
            key={i}
            size={size}
            className={filled || isHalf ? "fill-amber-400 text-amber-400" : "text-muted-foreground/40"}
            style={isHalf ? { clipPath: "inset(0 0 0 50%)" } : undefined}
          />
        );
      })}
    </span>
  );
}
