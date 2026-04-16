import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // ERP Audit Design System — "Professional Audit"
        primary: {
          DEFAULT: '#1F3A70',   // Deep navy
          foreground: '#FFFFFF',
        },
        secondary: {
          DEFAULT: '#D4A574',   // Gold accent
          foreground: '#1A1A1A',
        },
        background: '#F5F5F5',
        action: '#2E5090',      // CTA / main buttons
        success: '#2D6A4F',
        danger: '#9B2226',
        surface: '#FFFFFF',
        border: '#E0E0E0',
        'text-primary': '#1A1A1A',
        'text-secondary': '#5A5A5A',
      },
      fontFamily: {
        sans: ['Inter', 'Segoe UI', 'system-ui', 'sans-serif'],
        mono: ['Fira Code', 'monospace'],
      },
      borderRadius: {
        DEFAULT: '6px',
        card: '8px',
      },
      boxShadow: {
        card: '0 1px 3px rgba(0,0,0,0.08)',
      },
    },
  },
  plugins: [],
};

export default config;
