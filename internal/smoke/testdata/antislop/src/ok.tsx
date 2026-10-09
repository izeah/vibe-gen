// Clean code that must not trip anything.
export const Ok = () => (
  <main>
    <form>
      <label>
        Email
        <input name="email" type="email" />
      </label>
      <button type="submit">Sign in</button>
      <a href="#main">Skip to content</a>
      <p>&#8217; Could not reach the server. Check your connection and try again.</p>
    </form>
  </main>
);
