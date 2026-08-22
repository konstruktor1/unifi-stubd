## Summary

-

## Target Stage

- [ ] Topic branch -> `dev`
- [ ] `dev` -> `main`
- [ ] Hotfix branch -> `main`, then back to `dev`

## Checks

- [ ] Complete `make code-checker` gate
- [ ] `make package` through Build Runner when packaging or install layout changed
- [ ] `make integration-nonunifi-metrics` when adoption, metrics, or controller compatibility changed

## Safety

- [ ] No lab secrets, authkeys, private controller URLs, or PCAPs are included
- [ ] Controller commands do not mutate host networking without explicit code review
