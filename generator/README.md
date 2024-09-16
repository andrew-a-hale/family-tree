# Generator
Generate family trees on a generational level.

## Parameters
- Person

- Couple
  - Probability of having n children: Poisson Distribution with Lambda n

- Generation
  - Affine transformation of Lambda
  - Competition (c): range [0, 1] indication amount of competition
    - Is a seed value that oscillates by c by c^(3/2) or c^(2/3), and moves way
    from growth and declide
  - Generation Types:
    - Decline
    - Growth
    - Stable
