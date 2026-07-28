# COURSE-P2-T14 Screenshot Evidence

Status: PASSED

Captured through the in-app browser against the rebuilt local stack on
2026-07-27. Authentication used the approved local identity and the normal
Kratos login flow. No cookie, storage, or browser-profile inspection occurred.

| File | Viewport/state |
|---|---|
| `t14-desktop-enrollment-required.png` | 1280x900; backend enrollment-required response and real enrollment action |
| `t14-desktop-deep-list.png` | 1280x900; persisted depth-4 list with published item and draft excluded |
| `t14-desktop-rendered-detail.png` | 1280x900; ordered representative blocks, safe links, and backend Next |
| `t14-mobile-deep-list.png` | 360x800; responsive two-item API-ordered list |
| `t14-mobile-rendered-detail.png` | 360x800; responsive block rendering |
| `t14-mobile-empty-state.png` | 360x800; successful empty node |
| `t14-desktop-signed-out-denial.png` | 1280x900; final signed-out denial state |

The representative metadata and adjacent published item were temporary local
fixtures created through existing authenticated APIs. They were removed after
capture; the retained item was restored and database residue was zero.
