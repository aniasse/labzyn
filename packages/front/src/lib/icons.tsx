import {
  Box,
  FlaskConical,
  Folder,
  List,
  Database,
  FileText,
  BarChart3,
  Rocket,
  Settings,
  Boxes,
  Beaker,
  Layers,
  type LucideIcon,
} from 'lucide-react';

// Maps Twenty-style icon names (IconFolder, IconFlask…) to lucide icons.
const map: Record<string, LucideIcon> = {
  IconList: List,
  IconFolder: Folder,
  IconFlask: FlaskConical,
  IconBeaker: Beaker,
  IconDatabase: Database,
  IconFile: FileText,
  IconFileText: FileText,
  IconChartHistogram: BarChart3,
  IconRocket: Rocket,
  IconSettings: Settings,
  IconBox: Box,
  IconBoxes: Boxes,
  IconLayers: Layers,
};

/** Resolve a Twenty icon name to a lucide component, with a sane fallback. */
export function getIcon(name?: string): LucideIcon {
  return (name && map[name]) || Box;
}
