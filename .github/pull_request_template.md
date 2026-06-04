## Summary

-

## Target Stage

- [ ] Topic branch -> `dev`
- [ ] `dev` -> `main`
- [ ] Hotfix branch -> `main`, then back to `dev`

## Checks

- [ ] `make check`
- [ ] `make package` when packaging or install layout changed
- [ ] `make integration-docker` when inform, adoption, payloads, or controller compatibility changed

## Safety

- [ ] No lab secrets, authkeys, private controller URLs, or PCAPs are included
- [ ] Controller commands do not mutate host networking without explicit code review
