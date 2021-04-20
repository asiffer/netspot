---
title: Spot Algorithm
weight: 20
summary: Statistical learning algorithm behind netspot
---

SPOT is the core algorithm which monitors the 
network statistics. 
Its main strength is that it automatically provides a decision 
threshold based on the stream it monitors.

The threshold provided by Spot has the nice following features:
- Unsupervised computation (automatic, no label required)
- Probabilistic meaning (quantile)
- Stream ready computation (dynamic update, high throughput)

In particular this work has been [published](https://hal.archives-ouvertes.fr/hal-01640325/file/siffer_kdd_17.pdf) in 2017 at KDD conference:
> Siffer, A., Fouque, P. A., Termier, A., & Largouet, C. (2017, August). Anomaly detection in streams with extreme value theory. In Proceedings of the 23rd ACM SIGKDD International Conference on Knowledge Discovery and Data Mining (pp. 1067-1075). ACM

In Netspot, you can modify the Spot parameters either at 
global scale (all statistics) or specifically (statistics-wise)

### Probabilistic threshold

The parameter `q` is the main parameter of the Spot algorithm. It directly tunes the decision threshold 

$\mathbb{P}(X>z_q) = q$

### Model depth

### Calibration parameters

### Extra parameters