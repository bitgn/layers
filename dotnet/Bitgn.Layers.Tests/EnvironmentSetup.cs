using System.Threading;
using FoundationDB.Client;
using NUnit.Framework;

namespace Bitgn.Layers.Tests {
    [SetUpFixture]
    class EnvironmentSetup {
        readonly CancellationTokenSource _source = new CancellationTokenSource();
        IFdbDatabase _database;
        
        [OneTimeSetUp]
        public void RunBeforeAnyTests() {
            Fdb.Start();
            _database = Fdb.OpenAsync(_source.Token).GetAwaiter().GetResult();
            Env.Init(_database, _source.Token);
        }

        [OneTimeTearDown]
        public void RunAfterAnyTests() {
            using (_source)
            using (_database) { }
            
            Fdb.Stop();

        }
    }
}