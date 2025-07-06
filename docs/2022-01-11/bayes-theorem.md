---
layout: math
---

# Bayes Theorem

\$$\displaystyle P(A\vert B) = P(A)\cdot\frac{P(B\vert A)}{P(B)}\qquad\qquad(1)$$

If event $$A$$ means “have a disease”, and event $$B$$ means “have a positive test”, then, knowing that a person has a positive test (”given $$B$$“), how to calculate the probability $$P(A\vert B)$$ of this person having a disease (”probability of $$A$$ given $$B$$”) ?

Well, from the Bayes equation $$(1)$$, it’s the probability $$P(A)$$ of a random person (from the set of all people) having a disease (a.k.a. the “prior” probability), multiplied by $$\displaystyle \frac{P(B\vert A)}{P(B)}$$.

In practice, calculating $$\displaystyle \frac{P(B\vert A)}{P(B)}$$ is not straightforward — one needs to know $$P(B)$$, i. e. the total probability of getting a positive test.

Instead, if for the given test, we know two values: the *sensitivity* (true positive rate, or TPR) and *specificity* (true negative rate, or TNR), then it’s easier to use the equation $$(2)$$, which operates in terms of odds:

\$$\displaystyle O(A\vert B) = O(A)\cdot \frac{P(B\vert A)}{P(B\vert \neg A)}\qquad\qquad(2)$$

---

### Derivation of the Bayes equation $$(1)$$

If there’s an event $$A$$ in the universe $$U$$, then the probability of $$A$$ is given by: $$\displaystyle P(A) = \frac{\vert A\vert}{\vert U\vert}$$

For another event $$B$$, the probability is: $$\displaystyle P(B) = \frac{\vert B\vert}{\vert U\vert}$$

If events $$A$$ and $$B$$ intersect, e. g. we know we’re in the region $$B$$, what is the probability of being in the intersection ($$A$$ “given $$B$$”)? Well obviously it’s the size of the intersection divided by the size of $$B$$: $$\displaystyle P(A\vert B) = \frac{\vert A\cap B\vert}{\vert B\vert}$$

The same goes for probability of $$B$$ “given $$A$$”: $$\displaystyle P(B\vert A) = \frac{\vert A\cap B\vert}{\vert A\vert}$$

The numerator is the same, so this follows: $$\displaystyle P(A\vert B)\cdot\vert B\vert = P(B\vert A)\cdot\vert A\vert$$

Using the first two equations to substitute $$\vert A\vert$$ and $$\vert B\vert$$ with their probabilities in $$U$$, we get the Bayes equation $$(1)$$: $$\displaystyle P(A\vert B)\cdot P(B) = P(B\vert A)\cdot P(A)$$

### Derivation of the odds equation $$(2)$$

The equation $$(2)$$ directly follows from:

Odds of having a disease are: $$\displaystyle O(A) = \frac{\vert A\vert }{\vert \neg A\vert }$$ (e. g. when $$P(A) = 0.1$$, the odds are “1 to 9”, or $$\displaystyle \frac{1}{9}$$). Here, $$A$$ is the set of people having the disease and $$\neg A$$ is the set of people not having the disease.

Odds of having a disease given a positive test is then (decreasing sizes of the sets in both numerator and denominator by the corresponding probabilities of getting a positive test result):

\$$\displaystyle O(A\vert B) = \frac{\vert A\vert \cdot P(B\vert A)}{\vert \neg A\vert \cdot P(B\vert \neg A)} = O(A)\cdot \frac{P(B\vert A)}{P(B\vert \neg A)}$$

The second term is known as the Bayes factor, where $$P(B\vert A)$$ is the test *sensitivity*, and $$P(B\vert \neg A)$$ is the test’s false positive rate (FPR, or *1-specificity*). The first term $$O(A)$$ is known as the “prior” odds.

### Example usage of $$(2)$$

Say the prior odds of having a disease are 1 to 99 (i. e. the probability is 1%).

If a test has 90% sensitivity (TPR = 90%) and 90% specificity (FPR = 10%), then its Bayes factor is 9.

Meaning that having a positive test result updates the prior odds by the factor of 9, and the odds of having a disease given a positive test become 9 to 99.

### Synonyms

- *Sensitivity*, recall, true positive rate (TPR), probability of detection, hit rate, power

  \$$\displaystyle =\frac {TP}{P}$$

  (the number of true positives divided by number of positives)

- *Specificity*, true negative rate (TNR), selectivity

  \$$\displaystyle =\frac {TN}{N}$$

  (the number of true negatives divided by number of negatives)

- Precision, positive predictive value (PPV)

  \$$\displaystyle =\frac {TP}{PP}$$

  (the number of true positives divided by predicted positives)

---

Sources:

1. [3blue1brown: the Bayes rule](https://www.youtube.com/watch?v=lG4VkPoG3ko)
2. [Visualizing Bayes Theorem](https://oscarbonilla.com/2009/05/visualizing-bayes-theorem/)
3. [Precision and Recall](https://en.wikipedia.org/wiki/Precision_and_recall#Definition)
