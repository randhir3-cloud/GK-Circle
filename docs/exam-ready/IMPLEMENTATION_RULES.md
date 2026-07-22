# GK Circle Exam Ready Implementation Rules

These rules apply to EVERY ticket.

## Rule 1: No Scope Creep
Never modify files outside the ticket scope.

## Rule 2: Order of Reading
Read documentation before writing code. Read in order:
1. `IMPLEMENTATION_RULES.md`
2. `MASTER_PLAN.md`
3. `ROADMAP.md`
4. `CURRENT_STATUS.md`
5. `AGENT_HANDOFF.md`
6. `BOARD.md`
7. Current Ticket

## Rule 3: Search Before Creating
Never create:
- duplicate services
- duplicate hooks
- duplicate DTOs
- duplicate utilities
- duplicate components

Always extend existing code.

## Rule 4: Source of Truth
Backend is the source of truth. Do not duplicate validation logic in the frontend.

## Rule 5: API Contracts
Frontend must consume backend contracts exactly. Never invent API fields.

## Rule 6: Dark Mode
Every feature must support Dark Mode.

## Rule 7: Responsive Layouts
Every page must support:
- Desktop
- Tablet
- Mobile

## Rule 8: Code Quality
Every new component must use:
- Strict TypeScript
- Responsive UI
- Accessibility (ARIA labels, keyboard navigation)
- Reusable primitives

## Rule 9: No Placeholders
Never leave:
- TODO
- FIXME
- Placeholder code
- Mock logic

## Rule 10: Clean Shutdown
Before stopping:
- Commit changes
- Update documentation
- Write a detailed handoff
- Leave the repository buildable

## Rule 11: Architectural Changes
Never redesign architecture. If architecture changes are required, document them first in `DECISIONS.md` before implementation.

## Rule 12: Sequential Execution
One Ticket Only. Do not begin another ticket until the current ticket is Completed or Handed Off.
