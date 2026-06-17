# TENDR CONFIG & STATE TRACKER
# STATUS: UNRESOLVED ONLY

[[RELEASE_BLOCKERS_HIGH]]
- [x] #RB-H1 | Fix Dashboard status agar tidak hardcoded RUNNING
- [x] #RB-H2 | Fix Dashboard port agar tidak hardcoded

[[FEATURES_RELIABILITY_HIGH]]
- [x] #FE-R1 | Startup provider validation
- [x] #FE-R2 | Provider health monitoring

[[FEATURES_DEV_EXPERIENCE_HIGH]]
- [x] #FE-DX1 | CLI: tendr health
- [x] #FE-DX2 | CLI: tendr test --model <alias>
- [x] #FE-DX3 | CLI: tendr doctor
- [x] #FE-DX4 | CLI: tendr logs
- [x] #FE-DX5 | CLI: tendr start --dry-run

[[FEATURES_COST_VISIBILITY_HIGH]]
- [x] #FE-CV1 | CLI: tendr cost --explain
- [x] #FE-CV2 | Daily cost warning threshold
- [x] #FE-CV3 | Monthly spend projection

[[FEATURES_SETUP_EXPERIENCE_HIGH]]
- [ ] #FE-SE1 | Interactive setup wizard
- [ ] #FE-SE2 | Provider auto-detection
- [ ] #FE-SE3 | API key validation during setup

[[MODIFY_COST_SYSTEM]]
- [ ] #MD-CS1 | Fix token pricing calculation
- [ ] #MD-CS2 | Expand pricing coverage
- [ ] #MD-CS3 | Improve cost reporting transparency
- [ ] #MD-CS4 | Add explainable cost breakdown

[[MODIFY_CONFIGURATION]]
- [ ] #MD-CF1 | Write config to ~/.tendr/config.yaml with 0600 permission
- [ ] #MD-CF2 | Remove hardcoded runtime values
- [ ] #MD-CF3 | Ensure all documented config fields actually work

[[MODIFY_GATEWAY]]
- [ ] #MD-GW1 | Normalize all provider errors
- [ ] #MD-GW2 | Improve fallback explanations
- [ ] #MD-GW3 | Add request limits
- [ ] #MD-GW4 | Improve observability

[[MODIFY_CACHE]]
- [ ] #MD-CH1 | Replace pseudo-LRU eviction with proper LRU
- [ ] #MD-CH2 | Add background cleanup
- [ ] #MD-CH3 | Improve cache metrics

[[MODIFY_LOGGING]]
- [ ] #MD-LG1 | Use RFC3339 timestamps and match documentation format
- [ ] #MD-LG2 | Improve structured logging

[[MODIFY_TUI]]
- [ ] #MD-TU1 | Remove hardcoded values
- [ ] #MD-TU2 | Implement documented shortcuts
- [ ] #MD-TU3 | Show real provider status
- [ ] #MD-TU4 | Improve log readability

[[REMOVE_SCOPE]]
- [ ] #RM-SC1 | Defer Semantic & Embedding-based cache to V3+
- [ ] #RM-SC2 | Remove Enterprise-oriented & premature analytics features

[[SIMPLIFY_CODE]]
- [ ] #SM-CD1 | Reduce TUI complexity until data is reliable
- [ ] #SM-CD2 | Remove unused pricing_snapshots table
- [ ] #SM-CD3 | Remove dead code in gateway, CLI, and business logic

[[TECH_DEBT_HIGH]]
- [ ] #TD-H1 | Fix ConfigView dependency on Viper global state
- [ ] #TD-H2 | Remove hardcoded latency threshold & rate limiter values
- [ ] #TD-H3 | Remove hardcoded dashboard values
- [ ] #TD-H4 | Fix duplicated cache code path

[[TECH_DEBT_MEDIUM]]
- [ ] #TD-M1 | Add cache concurrency tests
- [ ] #TD-M2 | Add Gemini mock server tests
- [ ] #TD-M3 | Improve flaky tests

[[TECH_DEBT_LOW]]
- [ ] #TD-L1 | Cleanup unused code & improve internal comments
- [ ] #TD-L2 | Refactor TUI rendering

next task is:                                                                      
                                                                                      
   [[FEATURES_RELIABILITY_HIGH]]                                                      
   - [ ] #FE-R1 | Startup provider validation                                         
   - [ ] #FE-R2 | Provider health monitoring                                          
                                                                                      
   in @docs/TODO.md                                                                   
                                                                                      
   can you do it @generalist?  