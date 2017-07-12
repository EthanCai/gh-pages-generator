---
title: 如何搞定技术面试编码环节
tags:
  - coding
categories:
  - 软件开发
date: 2017-07-12 15:26:55
---

编码，是软件工程师必须掌握的技能。以下介绍的方法，帮助软件工程师更顺利的通过技术面试的编码环节。

{% asset_img cracking_the_coding_skills.png %}

**解决问题的步骤**

1. 听：仔细聆听问题描述，完整获取信息
2. 举例：注意特殊用例、边界场景
3. 快速解决：尽快找到解决方案，不一定要最优。想想最优方案是什么样子的，你的最终解决方案可能处于当前方案、最优方案之间
4. 优化：优化第一版解决方案
  - BUD优化方法
    - 瓶颈，Bottlenecks
    - 不必要的工作，Unnecessary work
    - 重复的工作，Duplicated work
  - Four Algorithm Approaches
    - **Pattern Matching**: What problems is this similar to?
    - **Simplify & Generalize**: Tweak and solve simpler problem.
    - **Base Case & Build**: Does it sound recursive-ish?
    - **Data Structure Brainstorm**: Try various data structures.
  - 或者尝试下面的方法
    - Look for any unused info
    - Use a fresh example
    - Solve it "incorrectly"
    - Make time vs. space tradeoff
    - Precompute or do upfront work
    - Try a hash table or another data structure
5. 重新审视解决方案，确保在编码前理解每个细节
6. 实现
  - Write beautiful code
    - Modularize your code from the beginning
    - Refactor to clean up anything that isn't beautiful
7. 测试
  - FIRST Analyze
    - What's it doing? Why?
    - Anything that looks wired?
    - Error hot spots
  - THEN use test cases
    - Small test cases
    - Edge cases
    - Bigger test cases
  - When you find bugs, fix them carefully.

**需要掌握的基础知识**

1. 数据结构：hash tables, linked lists, stacks, queues, trees, tries, graphs, vectors, heaps
2. 算法：quick sort, merge sort, binary search, breadth-first search, depth-first search
3. 基础概念：Big-O Time, Big-O Space, Recursion & Memoization, Probability, Bit Manipulation

# 参考

- [Cracking the Coding interview](https://www.slideshare.net/gayle2/cracking-the-coding-interview-college)
- [Architecture of Tech Interviews](https://www.slideshare.net/gayle2/architecture-of-tech-interviews)
- [在AWS面试是怎样一种体验](https://mp.weixin.qq.com/s?__biz=MzA4ODMwMDcxMQ==&mid=2650892256&idx=1&sn=71e0987c7c61ca25c58f586b60de3305)
- [Amazon Leadership Principles](https://www.amazon.jobs/principles)
