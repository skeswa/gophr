import React from 'react';

export default React.createClass({
  getPair: function() {
    return this.props.pair || [];
  },
  render: function() {
    return <div className="Subscriptions-Edit">
      <h1>Subscriptions Edit</h1>
    </div>;
  }
});
