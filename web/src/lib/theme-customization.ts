/*
Copyright (C) 2023-2026 ArcMux contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
/**
 * Theme customization constants and types.
 *
 * Lives in `lib/` (not `context/`) so it can be imported alongside the
 * provider without breaking React Fast Refresh boundaries.
 */

export const THEME_PRESETS = [
  {
    value: 'default',
    name: 'Ember',
    swatches: ['oklch(0.71 0.14 38)', 'oklch(0.65 0.12 145)'],
  },
  {
    // Inspired by Anthropic's official brand language: warm cream canvas
    // (#faf9f5) paired with clay/coral (#d97757) as the single accent.
    value: 'anthropic',
    name: 'Anthropic',
    swatches: ['oklch(0.984 0.005 95)', 'oklch(0.685 0.142 38)'],
  },
  {
    value: 'charcoal',
    name: 'Charcoal',
    swatches: ['oklch(0.2 0 0)', 'oklch(0.96 0 0)'],
  },
  {
    value: 'slate',
    name: 'Slate',
    swatches: ['oklch(0.45 0.01 260)', 'oklch(0.5 0.08 200)'],
  },
  {
    value: 'coral-reef',
    name: 'Coral Reef',
    swatches: ['oklch(0.65 0.2 20)', 'oklch(0.82 0.08 350)'],
  },
  {
    value: 'sage',
    name: 'Sage',
    swatches: ['oklch(0.55 0.08 145)', 'oklch(0.5 0.06 190)'],
  },
  {
    value: 'midnight',
    name: 'Midnight',
    swatches: ['oklch(0.55 0.15 270)', 'oklch(0.5 0.12 220)'],
  },
  {
    value: 'sunrise',
    name: 'Sunrise',
    swatches: ['oklch(0.72 0.16 75)', 'oklch(0.65 0.12 50)'],
  },
  {
    value: 'moss',
    name: 'Moss',
    swatches: ['oklch(0.5 0.1 140)', 'oklch(0.55 0.08 170)'],
  },
  {
    value: 'plum',
    name: 'Plum',
    swatches: ['oklch(0.55 0.16 310)', 'oklch(0.5 0.12 280)'],
  },
  {
    value: 'frost',
    name: 'Frost',
    swatches: ['oklch(0.55 0.12 240)', 'oklch(0.5 0.1 200)'],
  },
] as const

export type ThemePreset = (typeof THEME_PRESETS)[number]['value']
export type ThemeRadius = 'default' | 'none' | 'sm' | 'md' | 'lg' | 'xl'
export type ThemeScale = 'default' | 'sm' | 'lg' | 'xl'
export type ContentLayout = 'full' | 'centered'

/**
 * Font axis for the theme.
 *
 * - `default` — resolve at runtime from the active preset
 *   (see `PRESET_DEFAULT_FONT`).
 * - `sans` — humanist sans (Inter), the project's UI fallback.
 * - `serif` — editorial serif (Lora + CJK fallbacks).
 */
export type ThemeFont = 'default' | 'sans' | 'serif'

export type ResolvedThemeFont = Exclude<ThemeFont, 'default'>

export type ThemeCustomization = {
  preset: ThemePreset
  font: ThemeFont
  radius: ThemeRadius
  scale: ThemeScale
  contentLayout: ContentLayout
}

export const DEFAULT_THEME_CUSTOMIZATION: ThemeCustomization = {
  preset: 'default',
  font: 'default',
  radius: 'default',
  scale: 'default',
  contentLayout: 'full',
}

export const THEME_PRESET_VALUES = new Set(
  THEME_PRESETS.map((p) => p.value)
) as ReadonlySet<ThemePreset>

export const THEME_FONT_VALUES: ReadonlySet<ThemeFont> = new Set([
  'default',
  'sans',
  'serif',
])

export const THEME_RADIUS_VALUES: ReadonlySet<ThemeRadius> = new Set([
  'default',
  'none',
  'sm',
  'md',
  'lg',
  'xl',
])

export const THEME_SCALE_VALUES: ReadonlySet<ThemeScale> = new Set([
  'default',
  'sm',
  'lg',
  'xl',
])

export const CONTENT_LAYOUT_VALUES: ReadonlySet<ContentLayout> = new Set([
  'full',
  'centered',
])

export const THEME_COOKIE_KEYS = {
  preset: 'theme_preset',
  font: 'theme_font',
  radius: 'theme_radius',
  scale: 'theme_scale',
  contentLayout: 'theme_content_layout',
} as const

/**
 * Preset → default font mapping. Used by the provider to resolve the user's
 * `font: 'default'` preference against the active preset.
 *
 * Co-located with the preset registry so a preset's signature typography
 * is declared in one place. Presets not listed here fall back to the
 * `resolveThemeFont` default of `sans`.
 */
export const PRESET_DEFAULT_FONT: Partial<
  Record<ThemePreset, ResolvedThemeFont>
> = {
  default: 'sans',
  anthropic: 'serif',
}

/**
 * Resolve a user font preference + active preset into the concrete font that
 * should drive the DOM. Pure function so it's safe to call inside both the
 * effect that applies the attribute and the UI preview that hints at what
 * `default` will render as.
 */
export function resolveThemeFont(
  font: ThemeFont,
  preset: ThemePreset
): ResolvedThemeFont {
  if (font === 'default') {
    return PRESET_DEFAULT_FONT[preset] ?? 'sans'
  }
  return font
}