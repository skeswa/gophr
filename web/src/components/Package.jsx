import React from 'react';

export default React.createClass({
  getPair: function() {
    return this.props.pair || [];
  },
  render: function() {
    return <div className="Package">
      <h1>Package</h1>
    </div>;
  }
});
