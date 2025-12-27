# Fusion Frontend

RSS Reader frontend built with React 19 and modern web technologies.

## Tech Stack

- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite
- **Routing**: TanStack Router
- **UI Components**: shadcn/ui + Tailwind CSS
- **State Management**: Zustand
- **Package Manager**: pnpm

## Development

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

## Project Structure

```
src/
├── components/
│   ├── ui/           # shadcn/ui components
│   ├── layout/       # Layout components
│   ├── feed/         # Feed-related components
│   ├── item/         # Item-related components
│   └── settings/     # Settings components
├── routes/           # TanStack Router routes
│   ├── __root.tsx    # Root layout
│   └── index.tsx     # Main view
├── lib/
│   ├── api/          # API client
│   └── utils.ts      # Utility functions
├── store/
│   └── app.ts        # Zustand store
└── main.tsx          # Application entry point
```

## Development Status

Phase 2.1 (Project Initialization) - ✅ Complete

Next steps: Implement base infrastructure (Phase 2.2)
