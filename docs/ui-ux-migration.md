# UI/UX Migration Complete

All main screens now use the consistent 3-panel layout via `ScreenLayout` component.

## Components

- `ScreenLayout`: Enforces top bar → content → bottom help structure
- Enhanced `DataTable`: Full row highlighting, larger checkboxes
- Style system: Consistent colors, borders, and typography

## Screen Inventory

- ✅ Home: Platform selection with repo status
- ✅ Build: Build execution with dynamic actions
- ✅ CI Dashboard: Pipeline/jobs with log viewing
- ✅ Caches: Cache management
- ✅ Validations: Validation results
- ✅ Hints: Developer hints
- ✅ Hydration: Hydration execution

## Adding New Screens

Use `ScreenLayout` for all new screens:

```go
func (m *Model) View() string {
    layout := components.ScreenLayout{
        Breadcrumb: []string{"🚀 Home", "NewScreen"},
        Context:    "context here",
        Content:    m.renderContent(),
        ActionKeys: []components.KeyBinding{
            {Key: "key", Label: "action"},
        },
        NavKeys: []components.KeyBinding{
            {Key: "esc", Label: "back"},
            {Key: "q", Label: "quit"},
        },
        Width:  m.Width,
        Height: m.Height,
    }
    return layout.Render()
}
```

## Style Guidelines

- Use bordered boxes (Card component) for status/info sections
- Keep interactive areas (tables, lists) without borders
- Top bar: compact, single line with breadcrumb + abbreviated context
- Bottom help: context-aware actions on left, universal nav on right
