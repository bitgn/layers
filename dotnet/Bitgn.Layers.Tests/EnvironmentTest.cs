using System;
using System.Threading.Tasks;
using FoundationDB.Client;
using NUnit.Framework;

namespace Bitgn.Layers.Tests {
    public class EnvironmentTest {
       
        [Test]
        public async Task ConnectionExists() {
            using (var tx = Env.Db.BeginTransaction(FdbTransactionMode.ReadOnly, Env.Token)) {
                var version = await tx.GetReadVersionAsync();
                Assert.That(version, Is.Not.EqualTo(0));
            }
        }
    }
}