// Motion primitive inspired by ReactBits CountUp
// (https://reactbits.dev/components/count-up,
// source: https://github.com/DavidHDev/react-bits/blob/main/src/ts-default/Components/CountUp/CountUp.tsx).
// No extra npm dependencies (React + requestAnimationFrame only). Adapted:
// minimal value-tween for KPI totals, formats via caller, honors
// prefers-reduced-motion by rendering the final value instantly.
import { useEffect, useRef, useState } from 'react';

interface CountUpProps {
  value: number;
  format: (n: number) => string;
  durationMs?: number;
  className?: string;
}

const easeOutCubic = (t: number) => 1 - Math.pow(1 - t, 3);

export default function CountUp({ value, format, durationMs = 700, className = '' }: CountUpProps) {
  const [display, setDisplay] = useState(value);
  const fromRef = useRef(value);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      setDisplay(value);
      fromRef.current = value;
      return;
    }
    const from = fromRef.current;
    if (from === value) return;
    const start = performance.now();
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / durationMs);
      setDisplay(Math.round(from + (value - from) * easeOutCubic(t)));
      if (t < 1) {
        rafRef.current = requestAnimationFrame(tick);
      } else {
        fromRef.current = value;
      }
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
    };
  }, [value, durationMs]);

  return (
    <span className={className} aria-label={format(value)}>
      {format(display)}
    </span>
  );
}
