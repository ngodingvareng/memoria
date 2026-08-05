import type { Story } from '@/types/story';

export const dummyNotes = [
  {
    id: 1,
    date: 'August 12, 2036',
    time: '10:00',
    published: true,
    tags: ['science', 'food'],
    content: (
      <>
        <h1>Garlic bread with cheese: What the science tells us</h1>
        <p>
          For years parents have espoused the health benefits of eating garlic
          bread with cheese to their children, with the food earning such an
          iconic status in our culture that kids will often dress up as warm,
          cheesy loaf for Halloween.
        </p>
        <p>
          But a recent study shows that the celebrated appetizer may be linked
          to a series of rabies cases springing up around the country.
        </p>
      </>
    ),
  },
  {
    id: 2,
    date: 'August 10, 2036',
    time: '14:30',
    color: 'rose',
    published: true,
    tags: ['work', 'remote'],
    content: (
      <>
        <h1>The future of remote work</h1>
        <h2>A paradigm shift in how we collaborate</h2>
        <p>
          The pandemic fundamentally changed how organizations think about
          workplace flexibility. What started as a necessity has evolved into a
          strategic advantage for many companies.
        </p>
        <h3>Key benefits observed</h3>
        <p>
          Studies show that remote workers report higher job satisfaction and
          reduced commute stress, leading to improved work-life balance.
        </p>
      </>
    ),
  },
  {
    id: 3,
    date: 'August 8, 2036',
    time: '09:15',
    published: false,
    tags: ['programming', 'rust', 'learning'],
    content: (
      <>
        <h1>Learning a new programming language</h1>
        <p>
          Started exploring Rust today. The ownership model is quite different
          from what I'm used to in JavaScript and Python.
        </p>
        <h2>First impressions</h2>
        <p>
          The compiler is incredibly helpful with error messages. It feels like
          having a mentor looking over your shoulder.
        </p>
        <h3>Resources I'm using</h3>
        <p>
          The Rust Book is excellent. Also found some great YouTube tutorials
          that explain the borrow checker in simple terms.
        </p>
      </>
    ),
  },
  {
    id: 4,
    date: 'August 5, 2036',
    time: '16:45',
    color: 'amber',
    published: true,
    tags: ['meeting', 'project'],
    content: (
      <>
        <h1>Notes from today's meeting</h1>
        <h2>Project Alpha updates</h2>
        <p>
          Team agreed to push the deadline by two weeks. The new target is end
          of September.
        </p>
        <h2>Action items</h2>
        <p>
          Need to review the database schema before Friday. Also scheduled a
          follow-up call with the design team.
        </p>
        <h3>Open questions</h3>
        <p>
          Still waiting for clarification on the authentication flow. Will ping
          the security team tomorrow.
        </p>
      </>
    ),
  },
  {
    id: 5,
    date: 'August 3, 2036',
    time: '11:20',
    color: 'emerald',
    published: true,
    tags: ['recipe', 'cooking', 'ramen'],
    content: (
      <>
        <h1>Recipe: Homemade Ramen</h1>
        <p>
          Finally perfected my tonkotsu broth recipe after months of
          experimentation. The secret is simmering the pork bones for at least
          12 hours.
        </p>
        <h2>Ingredients</h2>
        <p>
          Pork bones, chicken feet, kombu, shiitake mushrooms, green onions,
          ginger, and garlic.
        </p>
        <h3>Tips</h3>
        <p>
          Keep the heat at a rolling boil to emulsify the fats. This is what
          gives tonkotsu its creamy texture.
        </p>
      </>
    ),
  },
  {
    id: 6,
    date: 'August 1, 2036',
    time: '08:00',
    published: false,
    tags: ['personal', 'reflection'],
    content: (
      <>
        <h1>Morning reflection</h1>
        <p>
          Woke up early today and watched the sunrise. There's something
          peaceful about the quiet hours before the world wakes up.
        </p>
        <h2>Goals for this month</h2>
        <p>
          Focus on health, finish reading two books, and spend more quality time
          with family.
        </p>
      </>
    ),
  },
  {
    id: 7,
    date: 'July 28, 2036',
    time: '19:30',
    published: true,
    tags: ['book', 'review', 'self-improvement'],
    content: (
      <>
        <h1>Book review: Atomic Habits</h1>
        <h2>Key takeaways</h2>
        <p>
          Small changes compound over time. The 1% improvement philosophy is
          more sustainable than trying to overhaul everything at once.
        </p>
        <h3>Favorite quote</h3>
        <p>
          "You do not rise to the level of your goals. You fall to the level of
          your systems."
        </p>
        <p>
          This really changed how I think about building habits. Focus on the
          process, not just the outcome.
        </p>
      </>
    ),
  },
  {
    id: 8,
    date: 'July 25, 2036',
    time: '13:00',
    published: false,
    tags: [],
    content: (
      <>
        <p>What is this</p>
      </>
    ),
  },
  {
    id: 9,
    date: 'July 22, 2036',
    time: '17:45',
    color: 'sky',
    published: true,
    tags: ['debugging', 'programming', 'websocket'],
    content: (
      <>
        <h1>Debugging session notes</h1>
        <h2>The problem</h2>
        <p>
          Memory leak in the WebSocket connection handler. The application was
          consuming 2GB of RAM after running for 24 hours.
        </p>
        <h3>Root cause</h3>
        <p>
          Event listeners were not being properly cleaned up when connections
          closed. Added cleanup logic in the disconnect handler.
        </p>
        <p>Memory usage now stable at around 200MB. Crisis averted!</p>
        <p>
          Lorem ipsum dolor sit amet, consectetur adipisicing elit. Dignissimos,
          ducimus tempore suscipit amet labore rerum odit alias fugit at
          recusandae consequuntur velit fuga eligendi, hic vel? Architecto
          maxime temporibus dignissimos.
        </p>
      </>
    ),
  },
  {
    id: 10,
    date: 'July 20, 2036',
    time: '10:30',
    published: false,
    tags: ['ideas', 'app', 'brainstorm'],
    content: (
      <>
        <h1>Ideas for the new app</h1>
        <p>
          Brainstorming session for the note-taking application. Want to make
          something that feels natural and fast.
        </p>
        <h2>Core features</h2>
        <p>
          Markdown support, real-time sync, offline mode, simple and clean UI.
        </p>
        <h3>Nice to have</h3>
        <p>
          Collaboration features, version history, export to PDF, and keyboard
          shortcuts for power users.
        </p>
      </>
    ),
  },
];

export const dummyStories: Story[] = [
  {
    user: {
      id: 1,
      name: 'ShadCN',
      username: '@shadcn',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: '@shadcn',
    },
    id: 10,
    date: 'July 20, 2036',
    time: '10:30',
    tags: ['ideas', 'app', 'brainstorm'],
    content: (
      <>
        <h1>Ideas for the new app</h1>
        <p>
          Brainstorming session for the note-taking application. Want to make
          something that feels natural and fast.
        </p>
        <h2>Core features</h2>
        <p>
          Markdown support, real-time sync, offline mode, simple and clean UI.
        </p>
        <h3>Nice to have</h3>
        <p>
          Collaboration features, version history, export to PDF, and keyboard
          shortcuts for power users.
        </p>
      </>
    ),
  },
  {
    user: {
      id: 1,
      name: 'ShadCN',
      username: '@shadcn',
      imageSrc: 'https://github.com/shadcn.png',
      imageAlt: '@shadcn',
    },
    id: 9,
    date: 'July 22, 2036',
    time: '17:45',
    color: 'sky',
    tags: ['debugging', 'programming', 'websocket'],
    content: (
      <>
        <h1>Debugging session notes</h1>
        <h2>The problem</h2>
        <p>
          Memory leak in the WebSocket connection handler. The application was
          consuming 2GB of RAM after running for 24 hours.
        </p>
        <h3>Root cause</h3>
        <p>
          Event listeners were not being properly cleaned up when connections
          closed. Added cleanup logic in the disconnect handler.
        </p>
        <p>Memory usage now stable at around 200MB. Crisis averted!</p>
        <p>
          Lorem ipsum dolor sit amet, consectetur adipisicing elit. Dignissimos,
          ducimus tempore suscipit amet labore rerum odit alias fugit at
          recusandae consequuntur velit fuga eligendi, hic vel? Architecto
          maxime temporibus dignissimos.
        </p>
      </>
    ),
  },
];
