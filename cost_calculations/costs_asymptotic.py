import numpy as np

### This script calculates the asymptotic constants in the exponent in the various theorems in our paper.
### For example, this script calculates the functions we plot in figures 1 and 4.
### This code restricts attention to the t = 1 case i.e. no servers collude.

def isprime(q):
    if q == 1:
        return False
    x = 2
    while x * x <= q:
        if q % x == 0:
            return False
        x += 1
    return True

# Binary entropy
def H(pr):
    if pr in [0, 1]:
        return 0
    ret = -pr * np.log2(pr) - (1-pr) * np.log2(1-pr)
    return ret

# p(theta), defined in definition A.1 and theorem A.2
def get_p(theta, d):
    coef = (theta - np.arange(d+1))[::-1]
    roots = np.roots(coef)
    roots = np.real(roots[np.isreal(roots)])
    roots = roots[roots > 0]
    assert len(roots) == 1, f"Expected exactly one positive real root for inputs {theta} and {d}, got {roots}"
    return roots[0]

# CntExp(theta, d), defined in section 4.1 and theorem A.2
def CntExp(theta, d):
    p = get_p(theta, d)
    return np.log2(sum([p**i for i in range(d+1)])) - theta * np.log2(p)

# CntExp_\leq(theta, d), defined in section 4.1 and theorem A.2
def CntExpLeq(theta, d):
    if theta <= d/2:
        return CntExp(theta, d)
    else:
        return np.log2(d+1)

### Our PIR efficiency metrics. This code restricts attention to the t = 1 case i.e. no servers collude.

def verify_params(theta, d, s, q):
    assert isprime(q), f"Field size {q} should be a prime"
    assert q >= s, f"Field size {q} should be at least the number of servers {s}"
    assert q >= d+1, f"Field size {q} should be strictly bigger than the individual degree {d}"
    assert theta >= 0 and theta <= d/2, f"Theta {theta} should be contained in [0, d/2] = [0, {d/2}]"

# outputs alpha such that the server storage (and preprocessing time) is n^{alpha + o(1)}
# use_finite_differences = True: results in theorem 3.2 (q = 2) and theorem 5.3
# use_finite_differences = False: results in theorem 5.4
def preprocessing_exponent(theta, d, s, q, use_finite_differences=True):
    verify_params(theta, d, s, q)
    if use_finite_differences:
        return np.log2(q)/CntExp(theta, d)
    else:
        return (np.log2(q) + CntExp(theta/s, d))/CntExp(theta, d)

# outputs beta such that the server time per query (and communication if information-theoretic) is n^{beta + o(1)}
# use_finite_differences = True: results in theorem 3.2 (q = 2) and theorem 5.3
# use_finite_differences = False: results in theorem 5.4
def server_online_time(theta, d, s, q, use_finite_differences=True):
    verify_params(theta, d, s, q)
    if use_finite_differences:
        return (s * H(theta/s) + theta * np.log2(d))/(s * CntExp(theta, d))
    else:
        return CntExp(theta/s, d)/CntExp(theta, d)
    
### Pre-existing schemes by Ghoshal, Li, Ma, Dai, and Shi (TCC 2025, ePrint:2024/765) and Beimel, Ishai, and Malkin (CRYPTO 2000)

def glmds_theorem_4_9(s, q):
    assert isprime(q), f"Field size {q} should be a prime"
    # As written, the scheme of GLMDS25 requires q >= s+1, but a modification of their scheme can work with just q >= s and achieve essentially the same parameters
    # This is by splitting the polynomial into homogeneous slices and handling each slice separately.
    assert q >= s, f"Field size {q} should be at least the number of servers {s}"
    beta = 2/np.log2(q) # server time n^{beta + o(1)} per query
    alpha = 1 + 2/np.log2(q) # storage n^{alpha + o(1)}
    return beta, alpha

def glmds_theorem_6_4(theta, s, q):
    assert isprime(q), f"Field size {q} should be a prime"
    assert q >= s, f"Field size {q} should be at least the number of servers {s}"
    beta = H(theta/s)/H(theta) # server time n^{beta + o(1)} per query
    alpha = (np.log2(q) + H(theta/s))/H(theta) # storage n^{alpha + o(1)}

    assert np.isclose(beta, server_online_time(theta, 1, s, q, use_finite_differences=False))
    assert np.isclose(alpha, preprocessing_exponent(theta, 1, s, q, use_finite_differences=False))

    return beta, alpha