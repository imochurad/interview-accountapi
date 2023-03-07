# Form3 Take Home Exercise. Implementation by Ihor Mochurad.

LinkedIn Profile: https://www.linkedin.com/in/ihormochurad/

This is my first ever attempt at writing GoLang code. 
In my past I was mostly a Java engineer, Java was my bread and butter for over 14 years now. 
That's not to say that I haven't worked with other languages over the years. 

The code that I have written in this exercise is not an idiomatic Go,
but an attempt of a Java engineer to code in Go.
However, I am quite certain that I would be able to cover the gray area very quickly.

Even if I am not considered for further stages of the interview, I would love to get some feedback.

P.S.: I acknowledge that one of requirements was to write tests that only interact with the fake API. 
However, IMO, the better code coverage is easier achieved via unit tests.
(acoounts_client.go line coverage sits at a little over 95% and that is thanks to UTs only)
On top of that, if you are writing a library you do not necessarily have access to the instance of a service.
Unit tests fail fast and provide a shorter feedback cycle than ITs.

Can the same code coverage be achieved with the ITs? Yes, of course!

Therefore, I have decided to showcase both testing approaches here.