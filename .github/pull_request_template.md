## Summary

-

## Target Stage

- [ ] Topic branch -> `dev`
- [ ] `dev` -> `main`
- [ ] Hotfix branch -> `main`, then back to `dev`

## Checks

- [ ] Run the complete `make verify` gate from the prepared consumer
- [ ] `make package` through Build Runner when packaging or install layout changed
- [ ] `make integration-nonunifi-metrics` when adoption, metrics, or controller compatibility changed

Leave blocked checks unchecked and record the owner, reason and next public
command from the provider handoff. Running a blocked check does not satisfy it.

## Safety

- [ ] No lab secrets, authkeys, private controller URLs, or PCAPs are included
- [ ] Controller commands do not mutate host networking without explicit code review
