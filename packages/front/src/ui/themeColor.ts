// Twenty's accent color names (from twenty-ui theme MainColorNames).
export type ThemeColor =
  | 'red'
  | 'ruby'
  | 'crimson'
  | 'tomato'
  | 'orange'
  | 'amber'
  | 'yellow'
  | 'lime'
  | 'grass'
  | 'green'
  | 'jade'
  | 'mint'
  | 'turquoise'
  | 'cyan'
  | 'sky'
  | 'blue'
  | 'iris'
  | 'violet'
  | 'purple'
  | 'plum'
  | 'pink'
  | 'bronze'
  | 'gold'
  | 'brown'
  | 'gray';

// Deterministic color for a string (mirrors Twenty's stable-color behavior).
const PALETTE: ThemeColor[] = [
  'green', 'turquoise', 'sky', 'blue', 'purple', 'pink', 'red', 'orange', 'yellow', 'gray',
];

export function colorForString(input: string): ThemeColor {
  let hash = 0;
  for (let i = 0; i < input.length; i++) hash = (hash * 31 + input.charCodeAt(i)) | 0;
  return PALETTE[Math.abs(hash) % PALETTE.length];
}
