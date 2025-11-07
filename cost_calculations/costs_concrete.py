import numpy as np
import itertools

def isprime(q):
    if q == 1:
        return False
    x = 2
    while x * x <= q:
        if q % x == 0:
            return False
        x += 1
    return True

### This script concretely calculates the efficiency costs of our protocols.
### For example, this script can be used to calculate the numbers we present in table 1.

### Notation
# D: total degree
# d: max individual degree
# m: number of variables
# q: field size
# s: number of servers
# homog: True if the DB polynomial is homogeneous (this gives you a slightly smaller DB size, but you can get away with a field of size S rather than S+1 now), False else

# We first use dynamic programming to compute Count(D, d, m) for all values D, d, m within some range
# This procedure is sketched for example in lemma B.1 of our paper

maxm = 50
maxd = 50
maxD = 50

assert maxd <= maxD, "maxd should be <= maxD"

count = np.full((maxD, maxd, maxm), 0, dtype=np.float64) # float64 to avoid overflow issues

# base cases

# d = 0 forces the monomial to be constant, for any D or m
for m in range(1, maxm):
    count[0, 0, m] = 1

# m = 1 case is straightforward
for D in range(maxD):
    for d in range(D, maxd): # in the single variable case, we have D <= d
        count[D, d, 1] = 1

# recursive step
for m in range(2, maxm):
    for d in range(1, maxd):
        for D in range(maxD):
            count[D, d, m] = 0
            # iterate through all possible powers for the first monomial
            # want r <= d and r <= D
            for r in range(min(d, D)+1):
                count[D, d, m] += count[D-r, d, m-1] # note that r <= d and r <= D

# number of elements in {0, 1, ..., d}^m of Hamming weight <= l
def HW(l, d, m):
    ret = 0
    for i in range(l+1):
        ret += count[i, 1, m] * (d**i)
    return ret

### Our PIR efficiency metrics. This code restricts attention to the t = 1 case i.e. no servers collude.

def verify_params(D, d, m, q, homog, s=None):
    assert isprime(q), f"Field size {q} should be a prime"
    assert m > 0, "Number of variables should be > 0"
    assert q > d, "Field size q --> individual degrees should all be <= q-1"
    assert D <= d * m, "Total degree should be at most individual degree * number of variables"
    if s is not None:
        if homog:
            assert q >= s, "In homogeneous setting, need field to have order >= s"
        else:
            assert q >= s+1, "In inhomogeneous setting, need field to have order >= s+1"

# Calculates the database size n (in bits) that can be accommodated for certain parameters
def db_size(D, d, m, q, homog=True):
    verify_params(D, d, m, q, homog)
    if homog:
        return int(np.floor(count[D, d, m] * np.log2(q))) # can pack log q database bits into one field element
    else: # in this case, we count the number of monomials of total degree at most D
        return int(np.floor(count[:(D+1), d, m].sum() * np.log2(q)))

# the communication cost of one server's answer (in bits)
# use_finite_differences = True: results in theorem 3.2 (q = 2) and theorem 5.3
# use_finite_differences = False: results in theorem 5.4
def server_online_time(D, d, m, q, s, homog=True, use_finite_differences=True):
    verify_params(D, d, m, q, homog, s=s)

    if use_finite_differences:
        if homog:
            L = [D]
        else:
            L = list(range(0, D+1))

        # communication is sum_{l in L} |HW_{l/s, d, m}| * log q
        ret = 0
        for l in L:
            ret += HW(int(np.floor(l/s)), d, m)
        return int(np.ceil(ret * np.log2(q)))
    else:
        # derivatives to order floor(D/s) inclusive
        maxD = int(np.floor(D/s))
        return int(np.ceil(count[:(maxD+1), d, m].sum() * np.log2(q)))

# the server storage (in bits)
# use_finite_differences = True: results in theorem 3.2 (q = 2) and theorem 5.3
# use_finite_differences = False: results in theorem 5.4
def preprocessing_cost(D, d, m, q, s, homog=True, use_finite_differences=True):
    verify_params(D, d, m, q, homog, s=s)

    # derivatives to order floor(D/s) inclusive, multiplying by q**m for precomputing all possible responses (if not using finite differences)
    maxD = int(np.floor(D/s))

    if use_finite_differences:
        return int(np.ceil((q**m) * np.log2(q))) # just the truth table of evaluations
    else:
        return int(np.ceil(count[:(maxD+1), d, m].sum() * (q**m) * np.log2(q))) # truth table for all derivatives up to order D/s
    
### Utils for unit testing

def brute_force_check(D, d, m):
    candidates = itertools.product(range(d+1), repeat=m)
    cnt = 0
    for cand in candidates:
        if sum(cand) == D:
            cnt += 1
    
    assert cnt == count[D, d, m], f"error for {D}, {d}, {m}"

def hook_check(D, d, m):
    assert m+1 < maxm
    RHS = count[D, d, m+1]
    LHS = 0
    for i in range(max(D-d, 0), D+1):
        LHS += count[i, d, m]
    assert LHS == RHS

### Unit tests
if __name__ == '__main__':
    print("Carrying out tests")

    for D in range(5):
        for d in range(5):
            for m in range(1, 5):
                brute_force_check(D, d, m)
                hook_check(D, d, m)

    # Some manual checks
    for d in range(3, maxd):
        assert count[3, d, 6] == 56, f"error for 3, {d}, 6" # number of 6-tuples adding to 3 --> 8C3

    for m in range(1, maxm):
        for d in range(maxd):
            for D in range(d*m+1, maxD):
                assert count[D, d, m] == 0, f"Error for {D}, {d}, {m}" # if total degree is > dm, no solutions

    # multilinear case
    assert count[5, 1, 7] == 21, f"Error for 5, 1, 7" # 7C5 = 21
